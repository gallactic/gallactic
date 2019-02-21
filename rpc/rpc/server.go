package rpcc

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/gallactic/gallactic/rpc/rpc/config"
	rpcConf "github.com/gallactic/gallactic/rpc/rpc/config"
	"github.com/gin-gonic/gin"
	"github.com/hyperledger/burrow/logging"
	cors "github.com/tommy351/gin-cors"
	graceful "gopkg.in/tylerb/graceful.v1"
)

const (
	killTime = 100 * time.Millisecond
)

func NewServeProcess(config *rpcConf.ServerConfig, logger *logging.Logger,
	servers ...Server) (*ServeProcess, error) {
	var spConfig rpcConf.ServerConfig
	// var spConfig
	if config != nil {
		spConfig = *config
	} else {
		return nil, fmt.Errorf("Nil passed as server configuration")
	}
	stopChannel := make(chan struct{}, 1)
	startListeners := make([]chan struct{}, 0)
	stopListeners := make([]chan struct{}, 0)
	sp := &ServeProcess{
		config:           &spConfig,
		servers:          servers,
		stopChan:         stopChannel,
		startListenChans: startListeners,
		stopListenChans:  stopListeners,
		srv:              nil,
		logger:           logger.WithScope("ServeProcess"),
	}
	return sp, nil
}

func (sp *ServeProcess) Start() error {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	config := sp.config

	address := config.Bind.Address
	port := config.Bind.Port
	listenAddress := address + ":" + fmt.Sprintf("%d", port)
	var listener net.Listener

	ch := newCORSMiddleware(config.CORS)
	router.Use(gin.Recovery(), logHandler(sp.logger), contentTypeMiddleware, ch)

	if port == 0 {
		return fmt.Errorf("0 is not a valid port.")
	}
	srv := &graceful.Server{
		Server: &http.Server{
			Handler: router,
		},
	}

	// Start the servers/handlers
	for _, s := range sp.servers {
		s.Start(config, router)
	}

	listen, lErr := net.Listen("tcp", listenAddress)
	if lErr != nil {
		return lErr
	}

	// setting TLS if enabled
	if config.TLS.TLS {
		// code below taken from previous rpc server, but it's not used
		// addr := srv.Addr
		// if addr == "" {
		// 	addr = ":https"
		// }
		tlsConfig := &tls.Config{}
		if tlsConfig.NextProtos == nil {
			tlsConfig.NextProtos = []string{"http/1.1"}
		}
		var tlsErr error
		tlsConfig.Certificates = make([]tls.Certificate, 1)
		tlsConfig.Certificates[0], tlsErr = tls.LoadX509KeyPair(config.TLS.CertPath, config.TLS.KeyPath)
		if tlsErr != nil {
			return tlsErr
		}

		listener = tls.NewListener(listen, tlsConfig)
	} else {
		listener = listen
	}
	sp.srv = srv
	sp.logger.InfoMsg(
		"Server started.",
		"address: ", sp.config.Bind.Address,
		"port: ", sp.config.Bind.Port,
	)

	for _, c := range sp.startListenChans {
		c <- struct{}{}
	}

	// start the serve routine
	go func() {
		sp.srv.Serve(listener)
		for _, s := range sp.servers {
			s.Shutdown(context.Background())
		}
	}()
	// Listen to the process stop event, it will call 'Stop'
	// on the graceful Server. This happens when someone
	// calls 'Stop' on the process.
	go func() {
		<-sp.stopChan
		sp.logger.InfoMsg("Close signal sent to server")
		sp.srv.Stop(killTime)
	}()
	// Listen to the servers stop event. It is triggered when
	// the server has been fully shut down.
	go func() {
		<-sp.srv.StopChan()
		sp.logger.InfoMsg("Server stop event fired. Good bye.")
		for _, c := range sp.stopListenChans {
			c <- struct{}{}
		}
	}()
	return nil
}

// Stop will release the port, process any remaining requests
// up until the timeout duration is passed, at which point it
// will abort them and shut down.
func (sp *ServeProcess) Shutdown(ctx context.Context) error {
	var err error
	for _, s := range sp.servers {
		sErr := s.Shutdown(ctx)
		if sErr != nil {
			err = sErr
		}
	}

	lChan := sp.StopEventChannel()
	sp.stopChan <- struct{}{}
	select {
	case <-lChan:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Get a start-event channel from the server. The start event
// is fired after the Start() function is called, and after
// the server has started listening for incoming connections.
// An error here .
func (sp *ServeProcess) StartEventChannel() <-chan struct{} {
	lChan := make(chan struct{}, 1)
	sp.startListenChans = append(sp.startListenChans, lChan)
	return lChan
}

// Get a stop-event channel from the server. The event happens
// after the Stop() function has been called, and after the
// timeout has passed. When the timeout has passed it will wait
// for confirmation from the http.Server, which normally takes
// a very short time (milliseconds).
func (sp *ServeProcess) StopEventChannel() <-chan struct{} {
	lChan := make(chan struct{}, 1)
	sp.stopListenChans = append(sp.stopListenChans, lChan)
	return lChan
}

func logHandler(logger *logging.Logger) gin.HandlerFunc {
	logger = logger.WithScope("ginLogHandler")
	return func(ctx *gin.Context) {
		path := ctx.Request.URL.Path

		ctx.Next()

		clientIP := ctx.ClientIP()
		method := ctx.Request.Method
		statusCode := ctx.Writer.Status()
		rawError := ctx.Errors.String()

		logger.Info.Log(
			"clientIP: ", clientIP,
			"statusCode: ", statusCode,
			"method: ", method,
			"path: ", path,
			"error: ", rawError,
		)
	}
}

func newCORSMiddleware(c config.CORS) gin.HandlerFunc {
	o := cors.Options{
		AllowCredentials: c.AllowCredentials,
		AllowHeaders:     c.AllowHeaders,
		AllowOrigins:     c.AllowOrigins,
		ExposeHeaders:    c.ExposeHeaders,
		MaxAge:           time.Duration(c.MaxAge),
	}
	return cors.Middleware(o)
}

func contentTypeMiddleware(ctx *gin.Context) {
	if ctx.Request.Method == "POST" && ctx.ContentType() != "application/json" {
		ctx.AbortWithError(415, fmt.Errorf("Media type not supported: "+ctx.ContentType()))
	} else {
		ctx.Next()
	}
}
