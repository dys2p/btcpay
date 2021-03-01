package btcpay

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type InvoiceStatus string

const (
	InvoiceNew        InvoiceStatus = "New"
	InvoiceProcessing               = "Processing"
	InvoiceExpired                  = "Expired"
	InvoiceInvalid                  = "Invalid"
	InvoiceSettled                  = "Settled"
)

// SpeedPolicy defines when an invoice is considered confirmed.
type SpeedPolicy string

const (
	HighSpeed      SpeedPolicy = "HighSpeed"      // consider the invoice confirmed when the payment transaction has >= 0 confirmations (i.e. as soon as it is visible on the blockchain)
	MediumSpeed                = "MediumSpeed"    // consider the invoice confirmed when the payment transaction has >= 1 confirmation
	LowMediumSpeed             = "LowMediumSpeed" // consider the invoice confirmed when the payment transaction has >= 2 confirmations
	LowSpeed                   = "LowSpeed"       // consider the invoice confirmed when the payment transaction has >= 6 confirmations
)

type Invoice struct {
	InvoiceRequest
	ID                   string `json:"id"`
	CheckoutLink         string `json:"checkoutLink"`
	CreatedTime          int    `json:"createdTime"`
	ExpirationTime       int    `json:"expirationTime"`
	MonitoringExpiration int    `json:"monitoringExpiration"`
	Status               string `json:"status"`
	AdditionalStatus     string `json:"additionalStatus"`
}

type InvoiceRequest struct {
	Amount          float64 `json:"amount"`
	Currency        string  `json:"currency"`
	InvoiceMetadata `json:"metadata,omitempty"`
	InvoiceCheckout `json:"checkout,omitempty"`
}

// Additional information around the invoice that can be supplied.
type InvoiceMetadata struct {
	OrderID string `json:"orderId,omitempty"` // if the invoice has been created through a payment request, this is usually the ID of the payment request
}

type InvoiceCheckout struct {
	SpeedPolicy       SpeedPolicy `json:"speedPolicy,omitempty"` // default: store setting
	PaymentMethods    []string    `json:"paymentMethods,omitempty"`
	ExpirationMinutes int         `json:"expirationMinutes,omitempty"` // refers to the "paid" state, when the transaction becomes visible on the blockchain
	MonitoringMinutes int         `json:"monitoringMinutes,omitempty"`
	PaymentTolerance  float64     `json:"paymentTolerance,omitempty"`
	RedirectURL       string      `json:"redirectURL,omitempty"`
	DefaultLanguage   string      `json:"defaultLanguage,omitempty"`
}

func (store *Store) CreateInvoice(request *InvoiceRequest) (*Invoice, error) {

	payload, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	resp, err := store.DoRequest(http.MethodPost, "invoices", bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("response status: %d: %s", resp.StatusCode, body)
	}

	var invoice = &Invoice{}
	return invoice, json.Unmarshal(body, invoice)
}

func (store *Store) GetInvoice(id string) (*Invoice, error) {

	resp, err := store.DoRequest(http.MethodGet, fmt.Sprintf("invoices/%s", id), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("response status: %d: %s", resp.StatusCode, body)
	}

	var invoice = &Invoice{}
	return invoice, json.Unmarshal(body, invoice)
}
