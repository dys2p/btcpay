package btcpay

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

var ErrUnauthenticated = errors.New("unauthenticated")
var ErrUnauthorized = errors.New("unauthorized")
var ErrBadRequest = errors.New("bad request")
var ErrNotFound = errors.New("not found")

type Store struct {
	API           *API   `json:"-"`
	ID            string `json:"id"`
	WebhookSecret string `json:"webhookSecret"`
}

// LoadStore unmarshals a json config file into a Store.
// If the file doesn't exist, it is created and an error is returned.
// If the file exists, then CheckAuth is called and the result is returned.
func LoadStore(api *API, jsonPath string) (*Store, error) {
	var store = &Store{
		API: api,
	}
	data, err := os.ReadFile(jsonPath)
	switch {
	case err == nil:
		if err := json.Unmarshal(data, store); err != nil {
			return nil, err
		}
		return store, store.CheckAuth()
	case os.IsNotExist(err):
		return nil, CreateStoreConfig(jsonPath)
	default:
		return nil, err
	}
}

// CreateStoreConfig creates an empty json config file with empty values and chmod 600, so someone can fill in easily.
// CreateStoreConfig always returns an error.
func CreateStoreConfig(jsonPath string) error {
	data, err := json.Marshal(&Store{})
	if err != nil {
		return err
	}
	if err := os.WriteFile(jsonPath, data, 0600); err != nil {
		return err
	}
	return fmt.Errorf("created empty config file: %s", jsonPath)
}

// CheckAuth checks authentication and authorization by performing bogus CreateInvoice and GetInvoice calls and checking the result.
// It returns ErrUnauthenticated, ErrUnauthorized or nil.
func (store *Store) CheckAuth() error {
	if _, err := store.CreateInvoice(nil); err != ErrBadRequest {
		return err
	}
	if _, err := store.GetInvoice("not-existing"); err != ErrNotFound {
		return err
	}
	return nil
}

func (store *Store) DoRequest(method string, path string, body io.Reader) (*http.Response, error) {

	req, err := http.NewRequest(
		method,
		fmt.Sprintf("%s/api/v1/stores/%s/%s", store.API.URI, store.ID, path),
		body,
	)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", fmt.Sprintf("token %s", store.API.UserAPIKey))
	req.Header.Add("Content-Type", "application/json")

	return (&http.Client{
		Timeout: 10 * time.Second,
	}).Do(req)
}
