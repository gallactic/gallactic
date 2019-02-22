package rpc

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	rpcConf "github.com/gallactic/gallactic/rpc/rpc/config"
	"github.com/gin-gonic/gin"
	graceful "gopkg.in/tylerb/graceful.v1"
)

// Codec is the coder and decoder
type Codec interface {
	EncodeBytes(interface{}) ([]byte, error)
	Encode(interface{}, io.Writer) error
	DecodeBytes(interface{}, []byte) error
	Decode(interface{}, io.Reader) error
}

type TmCodec struct {
}

type HttpService interface {
	Process(*http.Request, http.ResponseWriter)
}

type JSONRPCServer struct {
	service HttpService
	running bool
}

type JSONService struct {
	codec           Codec
	defaultHandlers map[string]RequestHandlerFunc
}

type RequestHandlerFunc func(request *RPCRequest, requester interface{}) (interface{}, int, error)

type RPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
	Id      string          `json:"id"`
}

// RPCResponse MUST follow the JSON-RPC specification for Response object
// reference: http://www.jsonrpc.org/specification#response_object
type RPCResponse interface {
	AssertIsRPCResponse() bool
}

// RPCResultResponse MUST NOT contain the error member if no error occurred
type RPCResultResponse struct {
	Result  interface{} `json:"result"`
	Id      string      `json:"id"`
	JSONRPC string      `json:"jsonrpc"`
}

// RPCErrorResponse MUST NOT contain the result member if an error occurred
type RPCErrorResponse struct {
	Error   *RPCError `json:"error"`
	Id      string    `json:"id"`
	JSONRPC string    `json:"jsonrpc"`
}

// RPCError MUST be included in the Response object if an error occurred
type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	// Note: Data is currently unused, and the data member may be omitted
	// Data  interface{} `json:"data"`
}

type Server interface {
	Start(*rpcConf.ServerConfig, *gin.Engine)
	Running() bool
	Shutdown(ctx context.Context) error
}

type ServeProcess struct {
	config           *rpcConf.ServerConfig
	servers          []Server
	stopChan         chan struct{}
	startListenChans []chan struct{}
	stopListenChans  []chan struct{}
	srv              *graceful.Server
}
