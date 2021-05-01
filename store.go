package btcpay

import (
	"net/http"
)

type Store interface {
	CheckInvoiceAuth() error
	CreateInvoice(req *InvoiceRequest) (*Invoice, error)
	CreatePaymentRequest(req *PaymentRequestRequest) (*PaymentRequest, error)
	GetInvoice(id string) (*Invoice, error)
	GetPaymentRequest(id string) (*PaymentRequest, error)
	GetServerStatus() (*ServerStatus, error)
	InvoiceCheckoutLink(id string)
	PaymentRequestLink(id string)
	ProcessWebhook(req *http.Request) (*InvoiceEvent, error)
}

type ServerStatus struct {
	Version                 string       `json:"version"`
	Onion                   string       `json:"onion"`
	SupportedPaymentMethods []string     `json:"supportedPaymentMethods"`
	FullySynched            bool         `json:"fullySynched"`
	SyncStatuses            []SyncStatus `json:"syncStatus"`
}

type SyncStatus struct {
	CryptoCode      string `json:"cryptoCode"`
	ChainHeight     int    `json:"chainHeight"`
	SyncHeight      int    `json:"syncHeight"`
	NodeInformation struct {
		Headers              int     `json:"headers"`
		Blocks               int     `json:"blocks"`
		VerificationProgress float64 `json:"verificationProgress"`
	} `json:"nodeInformation"`
}
