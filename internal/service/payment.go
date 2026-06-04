package service

import (
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"

	"qris-payment/internal/model"
)

type PaymentService struct {
	store *Store
}

func NewPaymentService(store *Store) *PaymentService {
	log.Printf("[PaymentService] Initialized (Simulation Only)")
	return &PaymentService{
		store: store,
	}
}

func (s *PaymentService) CreateQRISPayment(amount int64) (*model.CreatePaymentResponse, error) {
	orderID := fmt.Sprintf("SIM-%s", uuid.New().String()[:12])
	// Default base url fallback for localhost simulation
	qrString := fmt.Sprintf("http://localhost:3002/api/simulate/pay/%s", orderID)

	payment := &model.Payment{
		OrderID:    orderID,
		Amount:     amount,
		Status:     model.StatusPending,
		QRString:   qrString,
		QRImageURL: "",
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
		QRImageURL: "",
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
