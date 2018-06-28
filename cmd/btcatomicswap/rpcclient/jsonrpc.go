package rpcclient

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
)

// RPCErrorCode represents an error code to be used as a part of an RPCError
// which is in turn used in a JSON-RPC Response object.
//
// A specific type is used to help ensure the wrong errors aren't used.
type RPCErrorCode int

// RPCError represents an error that is used as a part of a JSON-RPC Response
// object.
type RPCError struct {
	Code    RPCErrorCode `json:"code,omitempty"`
	Message string       `json:"message,omitempty"`
}

// Guarantee RPCError satisifies the builtin error interface.
var _, _ error = RPCError{}, (*RPCError)(nil)

// Error returns a string describing the RPC error.  This satisifies the
// builtin error interface.
func (e RPCError) Error() string {
	return fmt.Sprintf("%d: %s", e.Code, e.Message)
}

// IsValidIDType checks that the ID field (which can go in any of the JSON-RPC
// requests, responses, or notifications) is valid.  JSON-RPC 1.0 allows any
// valid JSON type.  JSON-RPC 2.0 (which bitcoind follows for some parts) only
// allows string, number, or null, so this function restricts the allowed types
// to that list.  This function is only provided in case the caller is manually
// marshalling for some reason.    The functions which accept an ID in this
// package already call this function to ensure the provided id is valid.
func IsValidIDType(id interface{}) bool {
	switch id.(type) {
	case int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64,
		string,
		nil:
		return true
	default:
		return false
	}
}

// Request is a type for raw JSON-RPC 1.0 requests.  The Method field identifies
// the specific command type which in turns leads to different parameters.
// Callers typically will not use this directly since this package provides a
// statically typed command infrastructure which handles creation of these
// requests, however this struct it being exported in case the caller wants to
// construct raw requests for some reason.
type Request struct {
	Jsonrpc string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
	ID      interface{} `json:"id"`
}

// NewRequestWithPositionalParameters returns a new JSON-RPC 1.0 request object given the provided id,
// method, and parameters.  The parameters are marshalled into a json.RawMessage
// for the Params field of the returned request object.  This function is only
// provided in case the caller wants to construct raw requests for some reason.
//
// Typically callers will instead want to create a registered concrete command
// type with the NewCmd or New<Foo>Cmd functions and call the MarshalCmd
// function with that command to generate the marshalled JSON-RPC request.
func NewRequestWithPositionalParameters(id interface{}, method string, params []interface{}) (*Request, error) {
	if !IsValidIDType(id) {
		str := fmt.Sprintf("the id of type '%T' is invalid", id)
		return nil, errors.New(str)
	}

	rawParams := make([]json.RawMessage, 0, len(params))
	for _, param := range params {
		marshalledParam, err := json.Marshal(param)
		if err != nil {
			return nil, err
		}
		rawMessage := json.RawMessage(marshalledParam)
		rawParams = append(rawParams, rawMessage)
	}

	return &Request{
		Jsonrpc: "1.0",
		ID:      id,
		Method:  method,
		Params:  rawParams,
	}, nil
}

// NewRequestWithNamedParameters should be merged with NewRequestWithPositionalParameters
func NewRequestWithNamedParameters(id interface{}, method string, params interface{}) (*Request, error) {
	if !IsValidIDType(id) {
		str := fmt.Sprintf("the id of type '%T' is invalid", id)
		return nil, errors.New(str)
	}

	return &Request{
		Jsonrpc: "1.0",
		ID:      id,
		Method:  method,
		Params:  params,
	}, nil
}

// makeParams creates a slice of interface values for the given struct.
//It is useful for positional parameters only.
func makeParams(rt reflect.Type, rv reflect.Value) []interface{} {
	numFields := rt.NumField()
	params := make([]interface{}, 0, numFields)
	for i := 0; i < numFields; i++ {
		rtf := rt.Field(i)
		rvf := rv.Field(i)
		if rtf.Type.Kind() == reflect.Ptr {
			if rvf.IsNil() {
				break
			}
			rvf.Elem()
		}
		params = append(params, rvf.Interface())
	}

	return params
}

// MarshalCmd marshals the passed command to a JSON-RPC request byte slice that
// is suitable for transmission to an RPC server.  The provided command type
// must be a registered type.  All commands provided by this package are
// registered by default.
func MarshalCmd(id interface{}, cmd interface{}) ([]byte, error) {

	method, namedParameters, err := CmdMethod(cmd)
	if err != nil {
		return nil, err
	}

	// The provided command must not be nil.
	rv := reflect.ValueOf(cmd)
	if rv.IsNil() {
		str := "the specified command is nil"
		return nil, errors.New(str)
	}

	var rawCmd *Request
	if !namedParameters {
		// Create a slice of interface values in the order of the struct fields
		// while respecting pointer fields as optional params and only adding
		// them if they are non-nil.
		rt := reflect.TypeOf(cmd)
		params := makeParams(rt.Elem(), rv.Elem())
		rawCmd, err = NewRequestWithPositionalParameters(id, method, params)
	} else {
		rawCmd, err = NewRequestWithNamedParameters(id, method, cmd)
	}

	if err != nil {
		return nil, err
	}
	//marshal the final JSON-RPC request.
	return json.Marshal(rawCmd)

}
