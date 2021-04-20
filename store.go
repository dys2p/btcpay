package btcpay

import (
	"io"
	"net/http"
)

type Store interface {
	CheckAuth() error
	CreateInvoice(request *InvoiceRequest) (*Invoice, error)
	CreatePaymentRequest(request *PaymentRequestRequest) (*PaymentRequest, error)
	DoRequest(method string, path string, body io.Reader) (*http.Response, error)
	GetAPI() *API
	GetInvoice(id string) (*Invoice, error)
	GetPaymentRequest(id string) (*PaymentRequest, error)
	ProcessWebhook(r *http.Request) (*InvoiceEvent, error)
}
