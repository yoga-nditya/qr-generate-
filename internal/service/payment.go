package service

import (
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"

	"qris-payment/internal/model"
)

// PaymentService handles business logic for payment operations.
type PaymentService struct {
	store *Store
}

// NewPaymentService creates a new PaymentService with the given store.
func NewPaymentService(store *Store) *PaymentService {
	log.Println("[PaymentService] Initialized (Simulation Only)")
	return &PaymentService{store: store}
}

// CreateQRISPayment creates a new simulated QRIS payment with the given amount.
func (s *PaymentService) CreateQRISPayment(amount int64) (*model.CreatePaymentResponse, error) {
	orderID := fmt.Sprintf("SIM-%s", uuid.New().String()[:12])
	qrString := fmt.Sprintf("http://localhost:3002/api/simulate/pay/%s", orderID)

	payment := &model.Payment{
		OrderID:   orderID,
		Amount:    amount,
		Status:    model.StatusPending,
		QRString:  qrString,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.store.Save(payment); err != nil {
		log.Printf("[PaymentService] Failed to save payment: %v", err)
		return nil, fmt.Errorf("store error: %w", err)
	}

	log.Printf("[PaymentService] Payment created: order=%s", orderID)

	return &model.CreatePaymentResponse{
		OrderID:  orderID,
		Amount:   amount,
		QRString: qrString,
		Status:   string(model.StatusPending),
	}, nil
}

// UpdateStatus updates the payment status for the given order ID.
func (s *PaymentService) UpdateStatus(orderID, transactionID, status string) error {
	return s.store.UpdateStatus(orderID, transactionID, status)
}

// SavePayment persists a payment record to the store.
func (s *PaymentService) SavePayment(p *model.Payment) error {
	return s.store.Save(p)
}

// GetPayment retrieves a payment by order ID. Returns false if not found.
func (s *PaymentService) GetPayment(orderID string) (*model.Payment, bool) {
	return s.store.Get(orderID)
}

func (s *PaymentService) GetAllPayments() []*model.Payment {
	return s.store.All()
}
