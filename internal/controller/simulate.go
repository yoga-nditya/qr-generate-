package controller

import (
	"fmt"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"qris-payment/internal/model"
	"qris-payment/internal/service"
)

const simulatePaymentAmount = 10000

const alreadyPaidHTML = `<!DOCTYPE html>
<html lang="id">
<head>
  <meta charset="UTF-8">
  <title>Simulasi QRIS</title>
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <style>
    body { font-family: -apple-system, sans-serif; text-align: center; padding: 50px 20px; background: #ffffff; color: #0f172a; margin: 0; display: flex; align-items: center; justify-content: center; min-height: 100vh; box-sizing: border-box; }
    .card { max-width: 360px; width: 100%; border: 1px solid #f1f5f9; border-radius: 28px; padding: 32px; box-shadow: 0 20px 40px -15px rgba(0,0,0,0.05); }
    .icon { width: 64px; height: 64px; border-radius: 50%; background: #ecfdf5; color: #10b981; display: inline-flex; align-items: center; justify-content: center; font-size: 32px; font-weight: bold; margin-bottom: 20px; }
    h1 { font-size: 20px; font-weight: 800; margin: 0 0 8px 0; }
    p { font-size: 13px; color: #475569; margin: 0; line-height: 1.5; }
  </style>
</head>
<body>
  <div class="card">
    <div class="icon">V</div>
    <h1>Sudah Terbayar</h1>
    <p>Transaksi ini sudah diselesaikan sebelumnya.</p>
  </div>
</body>
</html>`

const paymentSuccessHTML = `<!DOCTYPE html>
<html lang="id">
<head>
  <meta charset="UTF-8">
  <title>Simulasi Sukses</title>
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <style>
    body { font-family: -apple-system, sans-serif; text-align: center; padding: 50px 20px; background: #ffffff; color: #0f172a; margin: 0; display: flex; align-items: center; justify-content: center; min-height: 100vh; box-sizing: border-box; }
    .card { max-width: 360px; width: 100%; border: 1px solid #f1f5f9; border-radius: 28px; padding: 32px; box-shadow: 0 20px 40px -15px rgba(0,0,0,0.05); }
    .icon { width: 64px; height: 64px; border-radius: 50%; background: #ecfdf5; color: #10b981; display: inline-flex; align-items: center; justify-content: center; font-size: 32px; font-weight: bold; margin-bottom: 20px; }
    h1 { font-size: 20px; font-weight: 800; margin: 0 0 8px 0; }
    p { font-size: 13px; color: #475569; margin: 0; line-height: 1.5; }
  </style>
</head>
<body>
  <div class="card">
    <div class="icon">V</div>
    <h1>Simulasi Berhasil</h1>
    <p>Pembayaran sebesar Rp 10.000 telah berhasil disimulasikan. Halaman di monitor komputer Anda akan segera dialihkan.</p>
  </div>
</body>
</html>`

type SimulateController struct {
	svc *service.PaymentService
	hub *service.Hub
}

func NewSimulateController(svc *service.PaymentService, hub *service.Hub) *SimulateController {
	return &SimulateController{svc: svc, hub: hub}
}

func (h *SimulateController) CreateSimulate(c *fiber.Ctx) error {
	orderID := fmt.Sprintf("SIM-%s", uuid.New().String()[:12])
	qrString := fmt.Sprintf("%s/api/simulate/pay/%s", c.BaseURL(), orderID)

	payment := &model.Payment{
		OrderID:   orderID,
		Amount:    simulatePaymentAmount,
		Status:    model.StatusPending,
		QRString:  qrString,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := h.svc.SavePayment(payment); err != nil {
		log.Printf("[SimulateController] Failed to save payment: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}

	log.Printf("[SimulateController] Payment created: order=%s qr=%s", orderID, qrString)

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"order_id":   orderID,
			"amount":     payment.Amount,
			"qr_string":  qrString,
			"status":     string(model.StatusPending),
			"created_at": payment.CreatedAt,
		},
	})
}

// PaySimulate handles GET and POST /api/simulate/pay/:order_id.
// GET returns an HTML confirmation page; POST returns JSON.
func (h *SimulateController) PaySimulate(c *fiber.Ctx) error {
	orderID := c.Params("order_id")

	payment, ok := h.svc.GetPayment(orderID)
	if !ok {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Order tidak ditemukan",
		})
	}

	if payment.Status != model.StatusPending {
		if c.Method() == fiber.MethodGet {
			c.Set(fiber.HeaderContentType, "text/html; charset=utf-8")
			return c.SendString(alreadyPaidHTML)
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Order sudah tidak dalam status pending",
		})
	}

	transactionID := fmt.Sprintf("TXN-%s", uuid.New().String()[:8])

	if err := h.svc.UpdateStatus(orderID, transactionID, "settlement"); err != nil {
		log.Printf("[SimulateController] Failed to update status: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}

	h.hub.Broadcast(orderID, model.WSMessage{
		Type:    "payment_update",
		OrderID: orderID,
		Status:  "success",
		Data: map[string]interface{}{
			"transaction_id":     transactionID,
			"transaction_status": "settlement",
			"payment_type":       "simulate",
			"gross_amount":       "10000.00",
		},
	})

	log.Printf("[SimulateController] Payment successful: order=%s txn=%s", orderID, transactionID)

	if c.Method() == fiber.MethodGet {
		c.Set(fiber.HeaderContentType, "text/html; charset=utf-8")
		return c.SendString(paymentSuccessHTML)
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"order_id":       orderID,
			"transaction_id": transactionID,
			"status":         "success",
			"amount":         simulatePaymentAmount,
			"paid_at":        time.Now(),
		},
	})
}

// GetSimulateStatus handles GET /api/simulate/status/:order_id.
func (h *SimulateController) GetSimulateStatus(c *fiber.Ctx) error {
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
