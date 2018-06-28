package rpcclient

import (
	"bytes"
	"encoding/hex"
	"encoding/json"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/wire"
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

// FutureDumpPrivKeyResult is a future promise to deliver the result of a
// DumpPrivKeyAsync RPC invocation (or an applicable error).
type FutureDumpPrivKeyResult chan *response

// Receive waits for the response promised by the future and returns the private
// key corresponding to the passed address encoded in the wallet import format
// (WIF)
func (r FutureDumpPrivKeyResult) Receive() (*btcutil.WIF, error) {
	res, err := receiveFuture(r)
	if err != nil {
		return nil, err
	}

	// Unmarshal result as a string.
	var privKeyWIF string
	err = json.Unmarshal(res, &privKeyWIF)
	if err != nil {
		return nil, err
	}

	return btcutil.DecodeWIF(privKeyWIF)
}

// DumpPrivKeyAsync returns an instance of a type that can be used to get the
// result of the RPC at some future time by invoking the Receive function on the
// returned instance.
//
// See DumpPrivKey for the blocking version and more details.
func (c *Client) DumpPrivKeyAsync(address btcutil.Address) FutureDumpPrivKeyResult {
	addr := address.EncodeAddress()
	cmd := NewGetPrivateKeysCmd(addr)
	return c.sendCmd(cmd)
}

// DumpPrivKey gets the private key corresponding to the passed address encoded
// in the wallet import format (WIF).
//
func (c *Client) DumpPrivKey(address btcutil.Address) (*btcutil.WIF, error) {
	return c.DumpPrivKeyAsync(address).Receive()
}

// GetPrivateKeysCmd defines the getprivatekeys JSON-RPC command.
type GetPrivateKeysCmd struct {
	Addresses []string
}

// NewGetPrivateKeysCmd returns a new instance which can be used to issue a
// getprivatekeys JSON-RPC command.
func NewGetPrivateKeysCmd(addresses ...string) *GetPrivateKeysCmd {
	return &GetPrivateKeysCmd{
		Addresses: addresses,
	}
}

// FutureGetFeeRateResult is a future promise to deliver the result of
// a GetFeeRateAsync RPC invocation (or an applicable error).
type FutureGetFeeRateResult chan *response

// Receive waits for the response promised by the future and returns a the feerate.
func (r FutureGetFeeRateResult) Receive() (feerate btcutil.Amount, err error) {
	res, err := receiveFuture(r)
	if err != nil {
		return
	}

	// Unmarshal result as a string.
	err = json.Unmarshal(res, &feerate)
	return
}

// GetFeeRateCmd defines the getfeerate RPC command.
type GetFeeRateCmd struct {
}

// NewGetFeeRateCmd returns a new instance which can be used to issue a
// getfeerate JSON-RPC command.
func NewGetFeeRateCmd() *GetFeeRateCmd {
	return &GetFeeRateCmd{}
}

// GetFeeRateAsync returns an instance of a type that can be used to
// get the result of the RPC at some future time by invoking the Receive
// function on the returned instance.
//
// See GetUnusedAddress for the blocking version and more details.
func (c *Client) GetFeeRateAsync() FutureGetFeeRateResult {
	cmd := NewGetFeeRateCmd()
	return c.sendCmd(cmd)
}

// GetFeeRate Returns the  current optimal fee rate per kilobyte, according to config settings(static/dynamic)returns the first unused address of the wallet,
func (c *Client) GetFeeRate() (btcutil.Amount, error) {
	return c.GetFeeRateAsync().Receive()
}

// FuturePayToResult is a future promise to deliver the result of
// a payto  RPC invocation (or an applicable error).
type FuturePayToResult chan *response

// Receive waits for the response promised by the future and returns the transaction  and wether or not it is complete ( signed).
func (r FuturePayToResult) Receive() (tx *wire.MsgTx, complete bool, err error) {
	rawResp, err := receiveFuture(r)
	if err != nil {
		return
	}
	var resp struct {
		Complete bool   `json:"complete"`
		FDinal   bool   ` json:"final"`
		Hex      string `json:"hex"`
	}

	// Unmarshal result
	err = json.Unmarshal(rawResp, &resp)
	if err != nil {
		return
	}
	complete = resp.Complete
	fundedTxBytes, err := hex.DecodeString(resp.Hex)
	if err != nil {
		return nil, false, err
	}
	tx = &wire.MsgTx{}
	err = tx.Deserialize(bytes.NewReader(fundedTxBytes))
	if err != nil {
		return nil, false, err
	}

	return
}

// PayToCmd defines the payto RPC command.
type PayToCmd struct {
	Destination string  `json:"destination"`
	Amount      float64 `json:"amount"`
	UnSigned    bool    `json:"unsigned"`
}

// NewPayToCmd returns a new instance which can be used to issue a
// payto JSON-RPC command.
func NewPayToCmd(destination btcutil.Address, amount btcutil.Amount, unsigned bool) *PayToCmd {
	return &PayToCmd{
		Destination: destination.EncodeAddress(),
		Amount:      amount.ToBTC(),
		UnSigned:    unsigned,
	}
}

// PayToAsync returns an instance of a type that can be used to
// get the result of the RPC at some future time by invoking the Receive
// function on the returned instance.
//
// See PayTo for the blocking version and more details.
func (c *Client) PayToAsync(destination btcutil.Address, amount btcutil.Amount, unsigned bool) FuturePayToResult {
	cmd := NewPayToCmd(destination, amount, unsigned)
	return c.sendCmd(cmd)
}

// PayTo returns a funded transaction
func (c *Client) PayTo(destination btcutil.Address, amount btcutil.Amount, unsigned bool) (tx *wire.MsgTx, complete bool, err error) {
	return c.PayToAsync(destination, amount, unsigned).Receive()
}

func init() {
	RegisterCmd("getunusedaddress", (*GetUnusedAddressCmd)(nil), false)
	RegisterCmd("getprivatekeys", (*GetPrivateKeysCmd)(nil), false)
	RegisterCmd("getfeerate", (*GetFeeRateCmd)(nil), false)
	RegisterCmd("payto", (*PayToCmd)(nil), true)
}
