package model

import "time"

type PaymentStatus string

const (
	StatusPending PaymentStatus = "pending"
	StatusSuccess PaymentStatus = "success"
	StatusFailed  PaymentStatus = "failed"
	StatusExpired PaymentStatus = "expired"
)

type Payment struct {
	OrderID       string        `json:"order_id"`
	Amount        int64         `json:"amount"`
	Status        PaymentStatus `json:"status"`
	QRString      string        `json:"qr_string"`
	QRImageURL    string        `json:"qr_image_url"`
	TransactionID string        `json:"transaction_id"`
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
}

type CreatePaymentResponse struct {
	OrderID    string `json:"order_id"`
	Amount     int64  `json:"amount"`
	QRString   string `json:"qr_string"`
	QRImageURL string `json:"qr_image_url"`
	Status     string `json:"status"`
}

type WebhookPayload struct {
	TransactionTime   string `json:"transaction_time"`
	TransactionStatus string `json:"transaction_status"`
	TransactionID     string `json:"transaction_id"`
	StatusMessage     string `json:"status_message"`
	StatusCode        string `json:"status_code"`
	SignatureKey      string `json:"signature_key"`
	PaymentType       string `json:"payment_type"`
	OrderID           string `json:"order_id"`
	MerchantID        string `json:"merchant_id"`
	GrossAmount       string `json:"gross_amount"`
	FraudStatus       string `json:"fraud_status"`
	Currency          string `json:"currency"`
}

type WSMessage struct {
	Type    string      `json:"type"`
	OrderID string      `json:"order_id"`
	Status  string      `json:"status"`
	Data    interface{} `json:"data,omitempty"`
}
