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
)

var ErrUnauthenticated = errors.New("unauthenticated")
var ErrUnauthorized = errors.New("unauthorized")
var ErrBadRequest = errors.New("bad request")
var ErrNotFound = errors.New("not found")

type ServerStore struct {
	API           *API   `json:"-"`
	ID            string `json:"id"`
	WebhookSecret string `json:"webhookSecret"`
}

// LoadServerStore unmarshals a json config file into a ServerStore.
// If the file doesn't exist, it is created and an error is returned.
func LoadServerStore(api *API, jsonPath string) (*ServerStore, error) {
	var store = &ServerStore{
		API: api,
	}
	data, err := os.ReadFile(jsonPath)
	switch {
	case err == nil:
		return store, json.Unmarshal(data, store)
	case os.IsNotExist(err):
		return nil, CreateStoreConfig(jsonPath)
	default:
		return nil, err
	}
}

// CreateStoreConfig creates an empty json config file with empty values and chmod 600, so someone can fill in easily.
// CreateStoreConfig always returns an error.
func CreateStoreConfig(jsonPath string) error {
	data, err := json.Marshal(&ServerStore{})
	if err != nil {
		return err
	}
	if err := os.WriteFile(jsonPath, data, 0600); err != nil {
		return err
	}
	return fmt.Errorf("created empty config file: %s", jsonPath)
}

// CheckInvoiceAuth checks authentication and authorization by performing bogus CreateInvoice and GetInvoice calls and checking the result.
// It returns ErrUnauthenticated, ErrUnauthorized or nil.
func (store *ServerStore) CheckInvoiceAuth() error {
	if _, err := store.CreateInvoice(nil); err != ErrBadRequest {
		return err
	}
	if _, err := store.GetInvoice("not-existing"); err != ErrNotFound {
		return err
	}
	return nil
}

func (store *ServerStore) CreateInvoice(request *InvoiceRequest) (*Invoice, error) {

	payload, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	resp, err := store.DoRequest(http.MethodPost, "invoices", bytes.NewBuffer(payload))
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

func (store *ServerStore) CreatePaymentRequest(request *PaymentRequestRequest) (*PaymentRequest, error) {

	payload, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	resp, err := store.DoRequest(http.MethodPost, "payment-requests", bytes.NewBuffer(payload))
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

func (store *ServerStore) DoRequest(method string, path string, body io.Reader) (*http.Response, error) {
	return store.API.DoRequest(method, fmt.Sprintf("stores/%s/%s", store.ID, path), body)
}

func (store *ServerStore) GetAPI() *API {
	return store.API
}

func (store *ServerStore) GetInvoice(id string) (*Invoice, error) {

	resp, err := store.DoRequest(http.MethodGet, fmt.Sprintf("invoices/%s", id), nil)
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

func (store *ServerStore) GetPaymentRequest(id string) (*PaymentRequest, error) {

	resp, err := store.DoRequest(http.MethodGet, fmt.Sprintf("payment-requests/%s", id), nil)
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

func (store *ServerStore) ProcessWebhook(r *http.Request) (*InvoiceEvent, error) {

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