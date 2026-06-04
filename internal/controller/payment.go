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
}

func NewPaymentController(svc *service.PaymentService) *PaymentController {
	return &PaymentController{svc: svc}
}

func (h *PaymentController) CreatePayment(c *fiber.Ctx) error {
	resp, err := h.svc.CreateQRISPayment(10000)
	if err != nil {
		log.Printf("[Controller] CreatePayment error: %v", err)
		return c.Status(500).JSON(fiber.Map{
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
		return c.Status(404).JSON(fiber.Map{
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

	service.GlobalHub.Register(orderID, c)
	defer func() {
		service.GlobalHub.Unregister(orderID, c)
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
		_, _, err := c.ReadMessage()
		if err != nil {
			log.Printf("[WS] Disconnected: order=%s", orderID)
			break
		}
	}
}
