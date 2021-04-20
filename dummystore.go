package btcpay

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

// DummyStore is designed for testing only. It supports invoices only and is not thread-safe.
type DummyStore struct {
	Invoices map[string]*Invoice
}

func NewDummyStore() *DummyStore {
	return &DummyStore{
		Invoices: make(map[string]*Invoice),
	}
}

func (*DummyStore) CheckAuth() error {
	return nil
}

func (s *DummyStore) CreateInvoice(request *InvoiceRequest) (*Invoice, error) {
	id := fmt.Sprintf("dummy-invoice-%d", time.Now().UnixNano())
	invoice := &Invoice{
		InvoiceRequest:       *request,
		ID:                   id,
		CheckoutLink:         "http://example.com",
		CreatedTime:          time.Now().Unix(),
		ExpirationTime:       time.Now().Unix() + int64(60*request.ExpirationMinutes),
		MonitoringExpiration: time.Now().Unix() + int64(60*request.MonitoringMinutes),
		Status:               InvoiceSettled,
	}
	s.Invoices[id] = invoice
	return invoice, nil
}

func (*DummyStore) CreatePaymentRequest(request *PaymentRequestRequest) (*PaymentRequest, error) {
	return nil, errors.New("not implemented")
}

func (*DummyStore) DoRequest(method string, path string, body io.Reader) (*http.Response, error) {
	return nil, errors.New("not implemented")
}

func (s *DummyStore) GetAPI() *API {
	return &API{
		URI: "https://example.com",
	}
}

func (s *DummyStore) GetInvoice(id string) (*Invoice, error) {
	invoice, ok := s.Invoices[id]
	if ok {
		return invoice, nil
	} else {
		return nil, errors.New("not found")
	}
}

func (*DummyStore) GetPaymentRequest(id string) (*PaymentRequest, error) {
	return nil, errors.New("not implemented")
}

func (*DummyStore) ProcessWebhook(r *http.Request) (*InvoiceEvent, error) {
	return nil, errors.New("not implemented")
}
