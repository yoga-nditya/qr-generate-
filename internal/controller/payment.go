package controller

import (
	"encoding/json"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"

	"qris-payment/internal/model"
	"qris-payment/internal/service"
)

type PaymentController struct {
	svc *service.PaymentService
	hub *service.Hub
}

func NewPaymentController(svc *service.PaymentService, hub *service.Hub) *PaymentController {
	return &PaymentController{svc: svc, hub: hub}
}

func (h *PaymentController) CreatePayment(c *fiber.Ctx) error {
	resp, err := h.svc.CreateQRISPayment(10000)
	if err != nil {
		log.Printf("[PaymentController] CreatePayment error: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    resp,
	})
}

func (h *PaymentController) GetStatus(c *fiber.Ctx) error {
	orderID := c.Params("order_id")
	payment, ok := h.svc.GetPayment(orderID)
	if !ok {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Order tidak ditemukan",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    payment,
	})
}

func (h *PaymentController) WebSocket(c *websocket.Conn) {
	orderID := c.Params("order_id")
	log.Printf("[WS] New connection: order=%s", orderID)

	h.hub.Register(orderID, c)
	defer func() {
		h.hub.Unregister(orderID, c)
		c.Close()
	}()

	if payment, ok := h.svc.GetPayment(orderID); ok {
		initMsg := model.WSMessage{
			Type:    "init",
			OrderID: orderID,
			Status:  string(payment.Status),
		}
		if data, err := json.Marshal(initMsg); err == nil {
			c.WriteMessage(websocket.TextMessage, data)
		}
	}

	for {
		if _, _, err := c.ReadMessage(); err != nil {
			log.Printf("[WS] Disconnected: order=%s", orderID)
			break
		}
	}
}
