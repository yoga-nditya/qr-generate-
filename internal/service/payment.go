package service

import (
	"crypto/sha512"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/coreapi"

	"qris-payment/internal/model"
)

type PaymentService struct {
	store  *Store
	client coreapi.Client
}

func NewPaymentService(store *Store) *PaymentService {
	var env midtrans.EnvironmentType
	if os.Getenv("MIDTRANS_ENV") == "production" {
		env = midtrans.Production
	} else {
		env = midtrans.Sandbox
	}

	client := coreapi.Client{}
	client.New(os.Getenv("MIDTRANS_SERVER_KEY"), env)

	log.Printf("[Midtrans] Initialized env=%s merchant=%s",
		os.Getenv("MIDTRANS_ENV"),
		os.Getenv("MIDTRANS_MERCHANT_ID"),
	)

	return &PaymentService{
		store:  store,
		client: client,
	}
}

func (s *PaymentService) CreateQRISPayment(amount int64) (*model.CreatePaymentResponse, error) {
	orderID := fmt.Sprintf("QRIS-%s", uuid.New().String()[:12])

	log.Printf("[Midtrans] Creating QRIS charge: order_id=%s amount=%d", orderID, amount)

	chargeReq := &coreapi.ChargeReq{
		PaymentType: coreapi.PaymentTypeQris,
		TransactionDetails: midtrans.TransactionDetails{
			OrderID:  orderID,
			GrossAmt: amount,
		},
		Qris: &coreapi.QrisDetails{
			Acquirer: "gopay",
		},
	}

	resp, err := s.client.ChargeTransaction(chargeReq)
	if err != nil {
		log.Printf("[Midtrans] ❌ Charge error: %v", err)
		return nil, fmt.Errorf("midtrans error: %v", err)
	}

	log.Printf("[Midtrans] ✅ Response: status=%s order=%s", resp.TransactionStatus, resp.OrderID)
	log.Printf("[Midtrans] QRString length=%d", len(resp.QRString))
	log.Printf("[Midtrans] Actions count=%d", len(resp.Actions))

	qrString := resp.QRString
	qrImageURL := ""

	for _, action := range resp.Actions {
		log.Printf("[Midtrans] Action: name=%s url=%s", action.Name, action.URL)
		if action.Name == "generate-qr-code" {
			qrImageURL = action.URL
		}
	}

	payment := &model.Payment{
		OrderID:    orderID,
		Amount:     amount,
		Status:     model.StatusPending,
		QRString:   qrString,
		QRImageURL: qrImageURL,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := s.store.Save(payment); err != nil {
		log.Printf("[Store] ❌ Save error: %v", err)
		return nil, fmt.Errorf("store error: %v", err)
	}

	log.Printf("[Store] ✅ Saved order=%s", orderID)

	return &model.CreatePaymentResponse{
		OrderID:    orderID,
		Amount:     amount,
		QRString:   qrString,
		QRImageURL: qrImageURL,
		Status:     string(model.StatusPending),
	}, nil
}

func (s *PaymentService) UpdateStatus(orderID, transactionID, status string) error {
	return s.store.UpdateStatus(orderID, transactionID, status)
}

func (s *PaymentService) SavePayment(p *model.Payment) error {
	return s.store.Save(p)
}

func (s *PaymentService) GetPayment(orderID string) (*model.Payment, bool) {
	return s.store.Get(orderID)
}

func (s *PaymentService) VerifySignature(orderID, statusCode, grossAmount, signatureKey string) bool {
	serverKey := os.Getenv("MIDTRANS_SERVER_KEY")
	raw := orderID + statusCode + grossAmount + serverKey
	h := sha512.New()
	h.Write([]byte(raw))
	computed := fmt.Sprintf("%x", h.Sum(nil))
	match := computed == signatureKey
	if !match {
		log.Printf("[Midtrans] ❌ Signature mismatch: order=%s", orderID)
	}
	return match
}
