package main

import (
	"log"

	"github.com/robvanmieghem/electrumatomicswap/cmd/btcatomicswap/rpcclient"
)

func main() {

	// Connect to local Electrum wallet  HTTP POST mode.
	connCfg := &rpcclient.ConnConfig{
		Host:         "localhost:7777",
		User:         "user",
		Pass:         "pass",
		HTTPPostMode: true, // Bitcoin core only supports HTTP POST mode
		DisableTLS:   true, // Bitcoin core does not provide TLS by default
	}

	client, err := rpcclient.New(connCfg)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Shutdown()
	// Get the current block count.
	addr, err := client.GetUnusedAddress()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Unused Address : %s", addr)
}
