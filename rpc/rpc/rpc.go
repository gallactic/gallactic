package rpcc

import (
	"encoding/json"
)

// JSON-RPC 2.0 error codes.
const (
	RPCErrorServerError    = -32000
	RPCErrorInvalidRequest = -32600
	RPCErrorMethodNotFound = -32601
	RPCErrorInvalidParams  = -32602
	RPCErrorInternalError  = -32603
	RPCErrorParseError     = -32700
)

// NewRPCRequest to Create a new RPC request. This is the generic struct that is passed to RPC methods
func NewRPCRequest(id string, method string, params json.RawMessage) *RPCRequest {
	return &RPCRequest{
		JSONRPC: "2.0",
		Id:      id,
		Method:  method,
		Params:  params,
	}
}

// NewRPCResponse creates a new response object from a result
func NewRPCResponse(id string, res interface{}) RPCResponse {
	return RPCResponse(&RPCResultResponse{
		Result:  res,
		Id:      id,
		JSONRPC: "2.0",
	})
}

// NewRPCErrorResponse creates a new error-response object from the error code and message
func NewRPCErrorResponse(id string, code int, message string) RPCResponse {
	return RPCResponse(&RPCErrorResponse{
		Error:   &RPCError{code, message},
		Id:      id,
		JSONRPC: "2.0",
	})
}

// AssertIsRPCResponse implements a marker method for RPCResultResponse
// to implement the interface RPCResponse
func (rpcResultResponse *RPCResultResponse) AssertIsRPCResponse() bool {
	return true
}

// AssertIsRPCResponse implements a marker method for RPCErrorResponse
// to implement the interface RPCResponse
func (rpcErrorResponse *RPCErrorResponse) AssertIsRPCResponse() bool {
	return true
}
