package btcpay

import (
	"encoding/json"
	"fmt"
	"os"
)

type API struct {
	URI           string `json:"uri"`           // without "/api" and without trailing slash
	UserAuthToken string `json:"userAuthToken"` // without "token" prefix, to be created in the user settings (not in the store settings)
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

func (api *API) PaymentRequestLink(id string) string {
	return fmt.Sprintf("%s/payment-requests/%s", api.URI, id)
}
