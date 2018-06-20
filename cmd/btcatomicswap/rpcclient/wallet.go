package rpcclient

import (
	"encoding/json"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
)

// FutureGetUnusedAddressResult is a future promise to deliver the result of
// a GetUnusedAddressAsync RPC invocation (or an applicable error).
type FutureGetUnusedAddressResult chan *response

// Receive waits for the response promised by the future and returns a new
// address.
func (r FutureGetUnusedAddressResult) Receive() (btcutil.Address, error) {
	res, err := receiveFuture(r)
	if err != nil {
		return nil, err
	}

	// Unmarshal result as a string.
	var addr string
	err = json.Unmarshal(res, &addr)
	if err != nil {
		return nil, err
	}

	return btcutil.DecodeAddress(addr, &chaincfg.MainNetParams)
}

// GetUnusedAddressCmd defines the getunusedaddress JSON-RPC command.
type GetUnusedAddressCmd struct {
}

// NewGetUnusedAddressCmd returns a new instance which can be used to issue a
// getunusedaddress JSON-RPC command.
func NewGetUnusedAddressCmd() *GetUnusedAddressCmd {
	return &GetUnusedAddressCmd{}
}

// GetUnusedAddressAsync returns an instance of a type that can be used to
// get the result of the RPC at some future time by invoking the Receive
// function on the returned instance.
//
// See GetUnusedAddress for the blocking version and more details.
func (c *Client) GetUnusedAddressAsync() FutureGetUnusedAddressResult {
	cmd := NewGetUnusedAddressCmd()
	return c.sendCmd(cmd)
}

// GetUnusedAddress returns the first unused address of the wallet,
// or None if all addresses are used.
// An address is considered as used if it has received a transaction, or if
//it is used in a payment request.
func (c *Client) GetUnusedAddress() (btcutil.Address, error) {
	return c.GetUnusedAddressAsync().Receive()
}

func init() {
	RegisterCmd("getunusedaddress", (*GetUnusedAddressCmd)(nil))
}
