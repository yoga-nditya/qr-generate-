package handler

import (
	"encoding/json"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"

	"qris-payment/internal/model"
	"qris-payment/internal/service"
)

type PaymentHandler struct {
	svc *service.PaymentService
}

func NewPaymentHandler(svc *service.PaymentService) *PaymentHandler {
	return &PaymentHandler{svc: svc}
}

func (h *PaymentHandler) IndexPage(c *fiber.Ctx) error {
	return c.Render("index", fiber.Map{
		"Title": "Pembayaran QRIS",
	})
}

func (h *PaymentHandler) CreatePayment(c *fiber.Ctx) error {
	resp, err := h.svc.CreateQRISPayment(1)
	if err != nil {
		log.Printf("[Handler] CreatePayment error: %v", err)
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

func (h *PaymentHandler) GetStatus(c *fiber.Ctx) error {
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

func (h *PaymentHandler) Webhook(c *fiber.Ctx) error {
	var payload model.WebhookPayload
	if err := c.BodyParser(&payload); err != nil {
		log.Printf("[Webhook] Parse error: %v", err)
		return c.Status(400).JSON(fiber.Map{"message": "bad request"})
	}

	log.Printf("[Webhook] 📩 Received: order_id=%s transaction_status=%s payment_type=%s",
		payload.OrderID, payload.TransactionStatus, payload.PaymentType)

	if !h.svc.VerifySignature(
		payload.OrderID,
		payload.StatusCode,
		payload.GrossAmount,
		payload.SignatureKey,
	) {
		return c.Status(401).JSON(fiber.Map{"message": "invalid signature"})
	}

	if err := h.svc.UpdateStatus(payload.OrderID, payload.TransactionID, payload.TransactionStatus); err != nil {
		log.Printf("[Webhook] Update error: %v", err)
	}

	wsStatus := mapStatus(payload.TransactionStatus)

	msg := model.WSMessage{
		Type:    "payment_update",
		OrderID: payload.OrderID,
		Status:  wsStatus,
		Data: map[string]interface{}{
			"transaction_id":     payload.TransactionID,
			"transaction_status": payload.TransactionStatus,
			"payment_type":       payload.PaymentType,
			"gross_amount":       payload.GrossAmount,
		},
	}
	service.GlobalHub.Broadcast(payload.OrderID, msg)

	return c.JSON(fiber.Map{"message": "ok"})
}

func (h *PaymentHandler) WebSocket(c *websocket.Conn) {
	orderID := c.Params("order_id")
	log.Printf("[WS] 🔌 New connection: order=%s", orderID)

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
			log.Printf("[WS] 🔌 Disconnected: order=%s", orderID)
			break
		}
	}
}

func mapStatus(midtransStatus string) string {
	switch midtransStatus {
	case "settlement", "capture":
		return "success"
	case "deny", "cancel", "failure":
		return "failed"
	case "expire":
		return "expired"
	default:
		return "pending"
	}
}
