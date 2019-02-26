package rpc

import (
	"context"

	rpcConf "github.com/gallactic/gallactic/www/rpc/config"
	"github.com/gin-gonic/gin"
)

// NewJSONServer to Create a new JSON RPC Server
func NewJSONServer(service HttpService) *JSONRPCServer {
	return &JSONRPCServer{service: service}
}

// Handler passes message on directly to the service which expects
// a normal http request and response writer.
func (jrs *JSONRPCServer) handleFunc(c *gin.Context) {
	r := c.Request
	w := c.Writer

	jrs.service.Process(r, w)
}

// Start adds the rpc path to the router.
func (jrs *JSONRPCServer) Start(config *rpcConf.ServerConfig, router *gin.Engine) {
	router.POST(config.HTTP.JsonRpcEndpoint, jrs.handleFunc)
}

// Running is to check the server currently running?
func (jrs *JSONRPCServer) Running() bool {
	return jrs.running
}

// Shutdown is to Shut the server down. Does nothing.
func (jrs *JSONRPCServer) Shutdown(ctx context.Context) error {
	jrs.running = false
	return nil
}
