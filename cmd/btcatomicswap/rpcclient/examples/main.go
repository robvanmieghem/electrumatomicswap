package main

import (
	"log"

	"github.com/btcsuite/btcutil"

	"github.com/robvanmieghem/electrumatomicswap/cmd/btcatomicswap/rpcclient"
)

func main() {

	// Connect to local Electrum wallet using HTTP POST mode.
	connCfg := &rpcclient.ConnConfig{
		Host:         "localhost:7777",
		User:         "user",
		Pass:         "pass",
		HTTPPostMode: true,
		DisableTLS:   true, // Electrum wallet does not provide TLS by default
	}

	client, err := rpcclient.New(connCfg)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Shutdown()
	// Get an unused address.
	addr, err := client.GetUnusedAddress()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Unused Address : %s", addr)

	feerate, err := client.GetFeeRate()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Feerate: %f BTC/KB", feerate.ToBTC())

	amount, err := btcutil.NewAmount(0.01)

	tx, complete, err := client.PayTo(addr, amount, true)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Transaction:", tx)
	log.Println("Complete:", complete)
}
