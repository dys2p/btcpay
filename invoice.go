package btcpay

const (
	InvoiceNew        string = "New"
	InvoiceProcessing string = "Processing"
	InvoiceExpired    string = "Expired"
	InvoiceInvalid    string = "Invalid"
	InvoiceSettled    string = "Settled"
)

// SpeedPolicy defines when an invoice is considered confirmed.
type SpeedPolicy string

const (
	HighSpeed      SpeedPolicy = "HighSpeed"      // consider the invoice confirmed when the payment transaction has >= 0 confirmations (i.e. as soon as it is visible on the blockchain)
	MediumSpeed    SpeedPolicy = "MediumSpeed"    // consider the invoice confirmed when the payment transaction has >= 1 confirmation
	LowMediumSpeed SpeedPolicy = "LowMediumSpeed" // consider the invoice confirmed when the payment transaction has >= 2 confirmations
	LowSpeed       SpeedPolicy = "LowSpeed"       // consider the invoice confirmed when the payment transaction has >= 6 confirmations
)

type Invoice struct {
	InvoiceRequest
	ID                   string `json:"id"`
	CheckoutLink         string `json:"checkoutLink"`
	CreatedTime          int64  `json:"createdTime"`
	ExpirationTime       int64  `json:"expirationTime"`
	MonitoringExpiration int64  `json:"monitoringExpiration"`
	Status               string `json:"status"`
	AdditionalStatus     string `json:"additionalStatus"`
}

type InvoiceRequest struct {
	Amount          float64 `json:"amount,string"`
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
