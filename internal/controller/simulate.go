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

type SimulateController struct {
	svc *service.PaymentService
}

func NewSimulateController(svc *service.PaymentService) *SimulateController {
	return &SimulateController{svc: svc}
}

func (h *SimulateController) CreateSimulate(c *fiber.Ctx) error {
	orderID := fmt.Sprintf("SIM-%s", uuid.New().String()[:12])

	// Dynamically use request base URL so it works seamlessly on local networks
	qrString := fmt.Sprintf("%s/api/simulate/pay/%s", c.BaseURL(), orderID)

	payment := &model.Payment{
		OrderID:    orderID,
		Amount:     10000,
		Status:     model.StatusPending,
		QRString:   qrString,
		QRImageURL: "",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := h.svc.SavePayment(payment); err != nil {
		log.Printf("[Simulate] Save error: %v", err)
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}

	log.Printf("[Simulate] ✅ Created order=%s (QR: %s)", orderID, qrString)

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

func (h *SimulateController) PaySimulate(c *fiber.Ctx) error {
	orderID := c.Params("order_id")

	payment, ok := h.svc.GetPayment(orderID)
	if !ok {
		return c.Status(404).JSON(fiber.Map{
			"success": false,
			"message": "Order tidak ditemukan",
		})
	}

	if payment.Status != model.StatusPending {
		if c.Method() == fiber.MethodGet {
			c.Set("Content-Type", "text/html; charset=utf-8")
			return c.SendString(`
				<!DOCTYPE html>
				<html>
				<head>
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
						<div class="icon">✓</div>
						<h1>Sudah Terbayar</h1>
						<p>Transaksi ini sudah diselesaikan sebelumnya.</p>
					</div>
				</body>
				</html>
			`)
		}
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "Order sudah tidak dalam status pending",
		})
	}

	transactionID := fmt.Sprintf("TXN-%s", uuid.New().String()[:8])

	if err := h.svc.UpdateStatus(orderID, transactionID, "settlement"); err != nil {
		log.Printf("[Simulate] Update error: %v", err)
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}

	msg := model.WSMessage{
		Type:    "payment_update",
		OrderID: orderID,
		Status:  "success",
		Data: map[string]interface{}{
			"transaction_id":     transactionID,
			"transaction_status": "settlement",
			"payment_type":       "simulate",
			"gross_amount":       "10000.00",
		},
	}
	service.GlobalHub.Broadcast(orderID, msg)

	log.Printf("[Simulate] ✅ Payment success order=%s txn=%s", orderID, transactionID)

	if c.Method() == fiber.MethodGet {
		c.Set("Content-Type", "text/html; charset=utf-8")
		return c.SendString(`
			<!DOCTYPE html>
			<html>
			<head>
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
					<div class="icon">✓</div>
					<h1>Simulasi Berhasil</h1>
					<p>Pembayaran sebesar Rp 10.000 telah berhasil disimulasikan. Halaman di monitor komputer Anda akan segera dialihkan.</p>
				</div>
			</body>
			</html>
		`)
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"order_id":       orderID,
			"transaction_id": transactionID,
			"status":         "success",
			"amount":         10000,
			"paid_at":        time.Now(),
		},
	})
}

func (h *SimulateController) GetSimulateStatus(c *fiber.Ctx) error {
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
