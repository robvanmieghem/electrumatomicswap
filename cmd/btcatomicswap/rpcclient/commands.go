package rpcclient

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/btcsuite/btcd/chaincfg/chainhash"

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
	var rawprivKeyWIF string
	err = json.Unmarshal(res, &rawprivKeyWIF)
	if err != nil {
		return nil, err
	}
	//Drop the "p2pkh:" prefix
	rawprivKeyWIF = strings.TrimPrefix(rawprivKeyWIF, "p2pkh:")
	return btcutil.DecodeWIF(rawprivKeyWIF)
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
	Address string
}

// NewGetPrivateKeysCmd returns a new instance which can be used to issue a
// getprivatekeys JSON-RPC command.
func NewGetPrivateKeysCmd(addresses ...string) (cmd *GetPrivateKeysCmd) {
	cmd = &GetPrivateKeysCmd{}
	for i, address := range addresses {
		if i != 0 {
			cmd.Address += ","
		}
		cmd.Address += address
	}
	return
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

//UnspentOutput represents an unspent output
type UnspentOutput struct {
	Address  btcutil.Address
	Value    btcutil.Amount
	OutPoint *wire.OutPoint
	Height   int64
}

// ListUnspentCmd defines the listunspent RPC command.
type ListUnspentCmd struct {
}

// NewListUnspentCmd returns a new instance which can be used to issue a
// getfeerate JSON-RPC command.
func NewListUnspentCmd() *ListUnspentCmd {
	return &ListUnspentCmd{}
}

// FutureListUnspentResult is a future promise to deliver the result of
// a listunspent RPC invocation (or an applicable error).
type FutureListUnspentResult chan *response

// Receive waits for the response promised by the future and returns the decode unspent outputs.
func (r FutureListUnspentResult) Receive() (utxos []*UnspentOutput, err error) {
	rawResp, err := receiveFuture(r)
	if err != nil {
		return
	}
	type respUtxo struct {
		Address     string `json:"address"`
		Value       string `json:"value"`
		PrevoutN    uint32 `json:"prevout_n"`
		PrevoutHash string `json:"prevout_hash"`
		Height      int64  `json:"height"`
		Coinbase    bool   `json:"coinbase"`
	}
	var resp []respUtxo
	// Unmarshal result
	err = json.Unmarshal(rawResp, &resp)
	if err != nil {
		return
	}
	utxos = make([]*UnspentOutput, len(resp))
	for i, respUtxo := range resp {
		utxo := &UnspentOutput{
			Height: respUtxo.Height,
		}
		value, err := strconv.ParseFloat(respUtxo.Value, 64)
		if err != nil {
			return nil, err
		}
		utxo.Value, err = btcutil.NewAmount(value)
		if err != nil {
			return nil, err
		}
		utxo.Address, err = btcutil.DecodeAddress(respUtxo.Address, nil)
		hash, err := chainhash.NewHashFromStr(respUtxo.PrevoutHash)
		if err != nil {
			return nil, err
		}
		utxo.OutPoint = wire.NewOutPoint(hash, respUtxo.PrevoutN)
		if err != nil {
			return nil, err
		}
		utxos[i] = utxo
	}

	return
}

// ListUnspentAsync returns an instance of a type that can be used to
// get the result of the RPC at some future time by invoking the Receive
// function on the returned instance.
//
// See ListUnspent for the blocking version and more details.
func (c *Client) ListUnspentAsync() FutureListUnspentResult {
	cmd := NewListUnspentCmd()
	return c.sendCmd(cmd)
}

//ListUnspent returns the list of unspent transaction outputs in the
//wallet by issuing a listunspent JSON-RPC command.
func (c *Client) ListUnspent() ([]*UnspentOutput, error) {
	return c.ListUnspentAsync().Receive()
}

// FutureBroadcastResult is a future promise to deliver the result of
// a broadcast RPC invocation (or an applicable error).
type FutureBroadcastResult chan *response

// Receive waits for the response promised by the future and returns a the feerate.
func (r FutureBroadcastResult) Receive() (*chainhash.Hash, error) {
	rawResponse, err := receiveFuture(r)
	var resp []interface{}
	err = json.Unmarshal(rawResponse, &resp)
	if err != nil {
		return nil, err
	}
	if len(resp) < 2 {
		return nil, fmt.Errorf("Invalid response: %s", string(rawResponse))
	}
	txID, ok := resp[1].(string)
	if !ok {
		return nil, fmt.Errorf("Invalid response: %s", string(rawResponse))
	}
	return chainhash.NewHashFromStr(txID)
}

// BroadcastCmd defines the  broadcast RPC command.
type BroadcastCmd struct {
	SerializedTransaction string
}

// NewBroadcastCmd returns a new instance which can be used to issue a
// broadcast JSON-RPC command.
func NewBroadcastCmd(tx *wire.MsgTx) (cmd *BroadcastCmd) {
	cmd = &BroadcastCmd{}
	var buf bytes.Buffer
	buf.Grow(tx.SerializeSize())
	err := tx.Serialize(&buf)
	if err != nil {
		//This should never happen
		panic(err)
	}

	cmd.SerializedTransaction = hex.EncodeToString(buf.Bytes())

	return
}

// BroadcastAsync returns an instance of a type that can be used to
// get the result of the RPC at some future time by invoking the Receive
// function on the returned instance.
//
// See Broadcast for the blocking version and more details.
func (c *Client) BroadcastAsync(tx *wire.MsgTx) FutureBroadcastResult {
	cmd := NewBroadcastCmd(tx)
	return c.sendCmd(cmd)
}

//Broadcast a transaction to the network
//by issuing a broadcast  JSON-RPC command
func (c *Client) Broadcast(tx *wire.MsgTx) (*chainhash.Hash, error) {
	return c.BroadcastAsync(tx).Receive()
}

//-----------------------
// Btc-Core compatibility
//-----------------------

// SendRawTransaction submits the encoded transaction to the server which will
// then relay it to the network.
//The allowHighFees parameter is ignored.
func (c *Client) SendRawTransaction(tx *wire.MsgTx, allowHighFees bool) (*chainhash.Hash, error) {
	return c.Broadcast(tx)
}

func init() {
	RegisterCmd("getunusedaddress", (*GetUnusedAddressCmd)(nil), false)
	RegisterCmd("getprivatekeys", (*GetPrivateKeysCmd)(nil), false)
	RegisterCmd("getfeerate", (*GetFeeRateCmd)(nil), false)
	RegisterCmd("payto", (*PayToCmd)(nil), true)
	RegisterCmd("listunspent", (*ListUnspentCmd)(nil), false)
	RegisterCmd("broadcast", (*BroadcastCmd)(nil), false)
}

//-----------------------
// Btc-Core compatibility
//-----------------------

// SignRawTransaction signs inputs for the passed transaction and returns the
// signed transaction as well as whether or not all inputs are now signed.
//
// This function assumes the RPC server already knows the input transactions and
// private keys for the passed transaction which needs to be signed and uses the
// default signature hash type.  Use one of the SignRawTransaction# variants to
// specify that information if needed.
func (c *Client) SignRawTransaction(tx *wire.MsgTx) (*wire.MsgTx, bool, error) {
	return nil, false, errors.New("SignRawTransaction is not implemented")
}
