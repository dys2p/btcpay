package btcpay

import (
	"time"
)

type PaymentRequestStatus string

const (
	PaymentRequestPending   PaymentRequestStatus = "Pending"
	PaymentRequestCompleted PaymentRequestStatus = "Completed"
	PaymentRequestExpired   PaymentRequestStatus = "Expired"
)

type PaymentRequest struct {
	PaymentRequestRequest
	Archived bool                 `json:"archived"`
	Created  string               `json:"created"`
	ID       string               `json:"id"`
	Status   PaymentRequestStatus `json:"status"`
}

// Mandatory fields are amount, currency and title.
type PaymentRequestRequest struct {
	AllowCustomPaymentAmounts bool    `json:"allowCustomPaymentAmounts,omitempty"`
	Amount                    float64 `json:"amount"`
	Currency                  string  `json:"currency"`                // ISO 4217 Currency code (BTC, EUR, USD, etc)
	CustomCSSLink             string  `json:"customCSSLink,omitempty"` // URI
	Description               string  `json:"description,omitempty"`   // HTML
	Email                     string  `json:"email,omitempty"`
	EmbeddedCSS               string  `json:"embeddedCSS,omitempty"` // CSS up to 500 bytes
	ExpiryDate                string  `json:"expiryDate,omitempty"`  // RFC3339 date (in contrast to the docs which say int64)
	Title                     string  `json:"title"`                 // required
}

func (req *PaymentRequestRequest) SetExpiryDays(days int) {
	req.ExpiryDate = time.Now().AddDate(0, 0, days).Format(time.RFC3339)
}
