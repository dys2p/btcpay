package btcpay

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type Store struct {
	API           *API   `json:"-"`
	ID            string `json:"id"`
	WebhookSecret string `json:"webhookSecret"`
}

// LoadStore unmarshals a json config file into a Store. If the file doesn't exist, it is created using CreateStoreConfig.
func LoadStore(api *API, jsonPath string) (*Store, error) {
	var store = &Store{
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
	data, err := json.Marshal(&Store{})
	if err != nil {
		return err
	}
	if err := os.WriteFile(jsonPath, data, 0600); err != nil {
		return err
	}
	return fmt.Errorf("created empty config file: %s", jsonPath)
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

	req.Header.Add("Authorization", fmt.Sprintf("token %s", store.API.UserAuthToken))
	req.Header.Add("Content-Type", "application/json")

	return (&http.Client{
		Timeout: 10 * time.Second,
	}).Do(req)
}
