package btcpay

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

var ErrUnauthenticated = errors.New("unauthenticated")
var ErrUnauthorized = errors.New("unauthorized")
var ErrBadRequest = errors.New("bad request")
var ErrNotFound = errors.New("not found")

type ServerStore struct {
	ServerURI     string `json:"uri"`        // without "/api" and without trailing slash, used for API access and for links to invoices and payment requests
	UserAPIKey    string `json:"userAPIKey"` // to be created in the BTCPay Server user settings (not in the store settings)
	ID            string `json:"id"`
	WebhookSecret string `json:"webhookSecret"`
}

// Load unmarshals a json config file into a ServerStore.
// If the file doesn't exist, it is created and an error is returned.
func Load(jsonPath string) (*ServerStore, error) {
	var store = &ServerStore{}
	data, err := os.ReadFile(jsonPath)
	switch {
	case err == nil:
		return store, json.Unmarshal(data, store)
	case os.IsNotExist(err):
		return nil, Create(jsonPath)
	default:
		return nil, err
	}
}

// Create creates an empty json config file with empty values and chmod 600, so someone can fill in easily.
// Create always returns an error.
func Create(jsonPath string) error {
	data, err := json.Marshal(&ServerStore{})
	if err != nil {
		return err
	}
	if err := os.WriteFile(jsonPath, data, 0600); err != nil {
		return err
	}
	return fmt.Errorf("created empty config file: %s", jsonPath)
}

func (s *ServerStore) doRequest(method string, path string, body io.Reader) (*http.Response, error) {

	req, err := http.NewRequest(
		method,
		fmt.Sprintf("%s/api/v1/%s", s.ServerURI, path),
		body,
	)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", fmt.Sprintf("token %s", s.UserAPIKey))
	req.Header.Add("Content-Type", "application/json")

	return (&http.Client{
		Timeout: 10 * time.Second,
	}).Do(req)
}

// CheckInvoiceAuth checks authentication and authorization by performing bogus CreateInvoice and GetInvoice calls and checking the result.
// It returns ErrUnauthenticated, ErrUnauthorized or nil.
func (s *ServerStore) CheckInvoiceAuth() error {
	if _, err := s.CreateInvoice(nil); err != ErrBadRequest {
		return err
	}
	if _, err := s.GetInvoice("not-existing"); err != ErrNotFound {
		return err
	}
	return nil
}

func (s *ServerStore) CreateInvoice(req *InvoiceRequest) (*Invoice, error) {

	payload, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := s.doRequest(http.MethodPost, fmt.Sprintf("stores/%s/invoices", s.ID), bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		// ok
	case http.StatusUnauthorized: // 401, "Unauthorized" should be "Unauthenticated"
		return nil, ErrUnauthenticated
	case http.StatusForbidden:
		return nil, ErrUnauthorized
	case http.StatusBadRequest:
		return nil, ErrBadRequest
	case http.StatusNotFound:
		return nil, ErrNotFound
	default:
		return nil, fmt.Errorf("response status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var invoice = &Invoice{}
	return invoice, json.Unmarshal(body, invoice)
}

func (s *ServerStore) CreatePaymentRequest(req *PaymentRequestRequest) (*PaymentRequest, error) {

	payload, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := s.doRequest(http.MethodPost, fmt.Sprintf("stores/%s/payment-requests", s.ID), bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		// ok
	case http.StatusUnauthorized: // 401, "Unauthorized" should be "Unauthenticated"
		return nil, ErrUnauthenticated
	case http.StatusForbidden:
		return nil, ErrUnauthorized
	case http.StatusBadRequest:
		return nil, ErrBadRequest
	case http.StatusNotFound:
		return nil, ErrNotFound
	default:
		return nil, fmt.Errorf("response status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var paymentRequest = &PaymentRequest{}
	return paymentRequest, json.Unmarshal(body, paymentRequest)
}

func (s *ServerStore) GetInvoice(id string) (*Invoice, error) {

	resp, err := s.doRequest(http.MethodGet, fmt.Sprintf("stores/%s/invoices/%s", s.ID, id), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		// ok
	case http.StatusUnauthorized: // 401, "Unauthorized" should be "Unauthenticated"
		return nil, ErrUnauthenticated
	case http.StatusForbidden:
		return nil, ErrUnauthorized
	case http.StatusBadRequest:
		return nil, ErrBadRequest
	case http.StatusNotFound:
		return nil, ErrNotFound
	default:
		return nil, fmt.Errorf("response status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var invoice = &Invoice{}
	return invoice, json.Unmarshal(body, invoice)
}

func (s *ServerStore) GetPaymentRequest(id string) (*PaymentRequest, error) {

	resp, err := s.doRequest(http.MethodGet, fmt.Sprintf("stores/%s/payment-requests/%s", s.ID, id), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		// ok
	case http.StatusUnauthorized: // 401, "Unauthorized" should be "Unauthenticated"
		return nil, ErrUnauthenticated
	case http.StatusForbidden:
		return nil, ErrUnauthorized
	case http.StatusBadRequest:
		return nil, ErrBadRequest
	case http.StatusNotFound:
		return nil, ErrNotFound
	default:
		return nil, fmt.Errorf("response status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var paymentRequest = &PaymentRequest{}
	return paymentRequest, json.Unmarshal(body, paymentRequest)
}

// GetServerStatus requires successful authentication, but no specific permissions.
func (s *ServerStore) GetServerStatus() (*ServerStatus, error) {

	resp, err := s.doRequest(http.MethodGet, "server/info", nil)
	if err != nil {
		return nil, err
	}

	switch resp.StatusCode {
	case http.StatusOK:
		// ok
	case http.StatusUnauthorized: // 401, "Unauthorized" should be "Unauthenticated"
		return nil, ErrUnauthenticated
	case http.StatusForbidden:
		return nil, ErrUnauthorized
	case http.StatusBadRequest:
		return nil, ErrBadRequest
	case http.StatusNotFound:
		return nil, ErrNotFound
	default:
		return nil, fmt.Errorf("response status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var status = &ServerStatus{}
	return status, json.Unmarshal(body, status)
}

func (s *ServerStore) InvoiceCheckoutLink(id string) string {
	return fmt.Sprintf("%s/i/%s", s.ServerURI, id)
}

func (s *ServerStore) PaymentRequestLink(id string) string {
	return fmt.Sprintf("%s/payment-requests/%s", s.ServerURI, id)
}

func (s *ServerStore) ProcessWebhook(r *http.Request) (*InvoiceEvent, error) {

	var messageMAC = []byte(strings.TrimPrefix(r.Header.Get("BTCPay-Sig"), "sha256="))
	if len(messageMAC) == 0 {
		return nil, errors.New("BTCPay-Sig header missing")
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	var mac = hmac.New(sha256.New, []byte(s.WebhookSecret))
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
	if event.StoreID != s.ID {
		return nil, fmt.Errorf("invoice store ID %s does not match selected store ID %s", event.StoreID, s.ID)
	}

	return event, err
}
