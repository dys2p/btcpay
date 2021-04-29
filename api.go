package btcpay

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type API struct {
	URI        string `json:"uri"`        // without "/api" and without trailing slash
	UserAPIKey string `json:"userAPIKey"` // to be created in the BTCPay Server user settings (not in the store settings)
}

// LoadAPI unmarshals a json config file into an API. If the file doesn't exist, it is created using CreateAPIConfig.
func LoadAPI(jsonPath string) (*API, error) {
	var api = &API{}
	data, err := os.ReadFile(jsonPath)
	switch {
	case err == nil:
		return api, json.Unmarshal(data, api)
	case os.IsNotExist(err):
		return nil, CreateAPIConfig(jsonPath)
	default:
		return nil, err
	}
}

// CreateAPIConfig creates an empty json config file with empty values and chmod 600, so someone can fill in easily.
// CreateAPIConfig always returns an error.
func CreateAPIConfig(jsonPath string) error {
	data, err := json.Marshal(&API{})
	if err != nil {
		return err
	}
	if err := os.WriteFile(jsonPath, data, 0600); err != nil {
		return err
	}
	return fmt.Errorf("created empty config file: %s", jsonPath)
}

func (api *API) DoRequest(method string, path string, body io.Reader) (*http.Response, error) {

	req, err := http.NewRequest(
		method,
		fmt.Sprintf("%s/api/v1/%s", api.URI, path),
		body,
	)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", fmt.Sprintf("token %s", api.UserAPIKey))
	req.Header.Add("Content-Type", "application/json")

	return (&http.Client{
		Timeout: 10 * time.Second,
	}).Do(req)
}

func (api *API) InvoiceCheckoutLink(id string) string {
	return fmt.Sprintf("%s/i/%s", api.URI, id)
}

func (api *API) PaymentRequestLink(id string) string {
	return fmt.Sprintf("%s/payment-requests/%s", api.URI, id)
}

// GetServerStatus requires no specific permissions.
func (api *API) GetServerStatus() (*ServerStatus, error) {

	resp, err := api.DoRequest(http.MethodGet, "server/info", nil)
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

type ServerStatus struct {
	Version                 string   `json:"version"`
	Onion                   string   `json:"onion"`
	SupportedPaymentMethods []string `json:"supportedPaymentMethods"`
	FullySynched            bool     `json:"fullySynched"`
	SyncStatus              []struct {
		CryptoCode      string `json:"cryptoCode"`
		ChainHeight     int    `json:"chainHeight"`
		SyncHeight      int    `json:"syncHeight"`
		NodeInformation struct {
			Headers              int     `json:"headers"`
			Blocks               int     `json:"blocks"`
			VerificationProgress float64 `json:"verificationProgress"`
		} `json:"nodeInformation"`
	} `json:"syncStatus"`
}
