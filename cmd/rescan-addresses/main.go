package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/dys2p/btcpay"
)

func main() {

	btcpayAPI, err := btcpay.LoadAPI("api.json")
	if err != nil {
		log.Printf("error loading btcpay API: %v", err)
		return
	}

	btcpayStore, err := btcpay.LoadStore(btcpayAPI, "store.json")
	if err != nil {
		log.Printf("error loading btcpay store: %v", err)
		return
	}

	for {
		fmt.Print(".")
		_, err := btcpayStore.CreateInvoice(&btcpay.InvoiceRequest{
			Amount:   0,
			Currency: "EUR",
			InvoiceCheckout: btcpay.InvoiceCheckout{
				ExpirationMinutes: 0,
				MonitoringMinutes: 0,
			},
		}) // CreateInvoice(nil) does not work, fails before trying to insert into the database
		if err == nil || !strings.Contains(err.Error(), "500") {
			break
		}
	}

	fmt.Println("\ndone")
}
