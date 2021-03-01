package btcpay

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type EventType string

// An invoice is considered "settled" if it has been "paid" in time (seen on the blockchain before the invoice time expired) and the full amount has been paid and the transaction has been confirmed (got n confirmations on the blockchain, see SpeedPolicy).
const (
	EventInvoiceCreated         EventType = "InvoiceCreated"
	EventInvoiceExpired                   = "InvoiceExpired"
	EventInvoiceInvalid                   = "InvoiceInvalid"
	EventInvoiceProcessing                = "InvoiceProcessing"
	EventInvoiceReceivedPayment           = "InvoiceReceivedPayment"
	EventInvoiceSettled                   = "InvoiceSettled"
)

type InvoiceEvent struct {
	DeliveryID         string    `json:"deliveryId"`
	InvoiceID          string    `json:"invoiceId"`
	IsRedelivery       bool      `json:"isRedelivery"`
	OriginalDeliveryID string    `json:"originalDeliveryId"`
	StoreID            string    `json:"storeId"`
	Timestamp          int64     `json:"timestamp"`
	Type               EventType `json:"type"`
	WebhookID          string    `json:"webhookId"`

	// InvoiceInvalid and InvoiceSettled only
	ManuallyMarked bool `json:"manuallyMarked"`

	// InvoiceReceivedPayment only
	AfterExpiration bool `json:"afterExpiration"` // whether this payment has been sent after the invoice expired

	// InvoiceExpired only
	PartiallyPaid bool `json:"partiallyPaid"`

	// InvoiceProcessing only
	OverPaid bool `json:"overPaid"`
}

func (store *Store) ProcessWebhook(r *http.Request) (*InvoiceEvent, error) {

	var messageMAC = []byte(strings.TrimPrefix(r.Header.Get("BTCPay-Sig"), "sha256="))
	if len(messageMAC) == 0 {
		return nil, errors.New("BTCPay-Sig header missing")
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	var mac = hmac.New(sha256.New, []byte(store.WebhookSecret))
	mac.Write(body)
	var expectedMAC = []byte(hex.EncodeToString(mac.Sum(nil)))
	if !hmac.Equal(messageMAC, expectedMAC) {
		return nil, fmt.Errorf("HMAC mismatch, got %s, want %s", messageMAC, expectedMAC)
	}

	var event = &InvoiceEvent{}
	if err := json.Unmarshal(body, event); err != nil {
		return nil, err
	}

	// mitigate BTCPayServer misconfigurations by checking the store ID
	if event.StoreID != store.ID {
		return nil, fmt.Errorf("invoice store ID %s does not match selected store ID %s", event.StoreID, store.ID)
	}

	return event, err
}
