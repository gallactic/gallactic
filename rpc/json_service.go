// Copyright 2017 Monax Industries Limited
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package rpc

import (
	"context"
	"encoding/json"
	"net/http"

	rpcConfig "github.com/gallactic/gallactic/rpc/config"
	"github.com/gin-gonic/gin"
	"github.com/hyperledger/burrow/logging"
)

// Server used to handle JSON-RPC 2.0 requests. Implements server.Server
type JsonRpcServer struct {
	service HttpService
	running bool
}

// Create a new JsonRpcServer
func NewJSONServer(service HttpService) *JsonRpcServer {
	return &JsonRpcServer{service: service}
}

// Start adds the rpc path to the router.
func (jrs *JsonRpcServer) Start(config *rpcConfig.ServerConfig, router *gin.Engine) {
	router.POST(config.HTTP.JsonRpcEndpoint, jrs.handleFunc)
	jrs.running = true
}

// Is the server currently running?
func (jrs *JsonRpcServer) Running() bool {
	return jrs.running
}

// Shut the server down. Does nothing.
func (jrs *JsonRpcServer) Shutdown(ctx context.Context) error {
	jrs.running = false
	return nil
}

// Handler passes message on directly to the service which expects
// a normal http request and response writer.
func (jrs *JsonRpcServer) handleFunc(c *gin.Context) {
	r := c.Request
	w := c.Writer

	jrs.service.Process(r, w)
}

// Used for Burrow. Implements server.HttpService
type JSONService struct {
	codec           Codec
	service         *Service
	defaultHandlers map[string]RequestHandlerFunc
	logger          *logging.Logger
}

// Create a new JSON-RPC 2.0 service for burrow (tendermint).
func NewJSONService(codec Codec, service *Service, logger *logging.Logger) HttpService {

	httpService := &JSONService{
		codec:   codec,
		service: service,
		logger:  logger.WithScope("NewJSONService"),
	}

	dhMap := GetMethods(codec, service)
	httpService.defaultHandlers = dhMap
	return httpService
}

// Process a request.
func (js *JSONService) Process(r *http.Request, w http.ResponseWriter) {

	// Create new request object and unmarshal.
	req := &RPCRequest{}
	decoder := json.NewDecoder(r.Body)
	errU := decoder.Decode(req)

	// Error when decoding.
	if errU != nil {
		js.writeError("Failed to parse request: "+errU.Error(), "",
			RPCErrorParseError, w)
		return
	}

	// Wrong protocol version.
	if req.JSONRPC != "2.0" {
		js.writeError("Wrong protocol version: "+req.JSONRPC, req.Id,
			RPCErrorInvalidRequest, w)
		return
	}

	mName := req.Method

	if handler, ok := js.defaultHandlers[mName]; ok {
		js.logger.TraceMsg("Request received",
			"id", req.Id,
			"method", req.Method)
		resp, errCode, err := handler(req, w)
		if err != nil {
			js.writeError(err.Error(), req.Id, errCode, w)
		} else {
			js.writeResponse(req.Id, resp, w)
		}
	} else {
		js.writeError("Method not found: "+mName, req.Id, RPCErrorMethodNotFound, w)
	}
}

// Helper for writing error responses.
func (js *JSONService) writeError(msg, id string, code int, w http.ResponseWriter) {
	response := NewRPCErrorResponse(id, code, msg)
	err := js.codec.Encode(response, w)
	// If there's an error here all bets are off.
	if err != nil {
		http.Error(w, "Failed to marshal standard error response: "+err.Error(), 500)
		return
	}
	w.WriteHeader(200)
}

// Helper for writing responses.
func (js *JSONService) writeResponse(id string, result interface{}, w http.ResponseWriter) {
	response := NewRPCResponse(id, result)
	err := js.codec.Encode(response, w)
	if err != nil {
		js.writeError("Internal error: "+err.Error(), id, RPCErrorInternalError, w)
		return
	}
	w.WriteHeader(200)
}
