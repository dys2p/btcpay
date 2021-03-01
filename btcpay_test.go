package btcpay

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"testing"
)

func TestInvoice(t *testing.T) {

	api, err := LoadAPI("api.json")
	if err != nil {
		t.Fatal(err)
	}

	store, err := LoadStore(api, "store.json")
	if err != nil {
		t.Fatal(err)
	}

	created, err := store.CreateInvoice(&InvoiceRequest{Amount: 1.23, Currency: "EUR"})
	if err != nil {
		t.Fatal(err)
	}

	got, err := store.GetInvoice(created.ID)
	if err != nil {
		t.Fatal(err)
	}

	if got.Amount != 1.23 || got.Currency != "EUR" {
		t.Fail()
	}

	// simulate a webhook

	var body = []byte(fmt.Sprintf(`
	{
		"deliveryId": "delivery-abcdefghijkl",
		"webhookId": "webhook-abcdefghijklmn",
		"orignalDeliveryId": "delivery-abcdefghijkl",
		"isRedelivery": false,
		"type": "InvoiceCreated",
		"timestamp": 1610000000,
		"storeId": "%s",
		"invoiceId": "%s"
	}`, store.ID, got.ID))

	var webhookRequest = &http.Request{
		Body:   io.NopCloser(bytes.NewReader(body)),
		Header: make(http.Header),
	}
	var mac = hmac.New(sha256.New, []byte(store.WebhookSecret))
	mac.Write(body)
	webhookRequest.Header.Add("BTCPay-Sig", fmt.Sprintf("sha256=%s", hex.EncodeToString(mac.Sum(nil))))

	event, err := store.ProcessWebhook(webhookRequest)
	if err != nil {
		t.Fatal(err)
	}

	if event.StoreID != store.ID || event.Type != EventInvoiceCreated || event.InvoiceID != got.ID {
		t.Fail()
	}
}

func Test1PaymentRequest(t *testing.T) {

	api, err := LoadAPI("api.json")
	if err != nil {
		t.Fatal(err)
	}

	store, err := LoadStore(api, "store.json")
	if err != nil {
		t.Fatal(err)
	}

	request := &PaymentRequestRequest{
		Amount:   1.23,
		Currency: "EUR",
		Title:    "My payment request",
	}
	request.SetExpiryDays(7)

	created, err := store.CreatePaymentRequest(request)
	if err != nil {
		t.Fatal(err)
	}

	got, err := store.GetPaymentRequest(created.ID)
	if err != nil {
		t.Fatal(err)
	}

	if got.Amount != 1.23 || got.Currency != "EUR" {
		t.Fail()
	}
}
