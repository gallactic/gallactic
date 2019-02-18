package grpc

import (
	"context"
	"fmt"
	"net"
	"runtime/debug"

	log "github.com/inconshreveable/log15"
	"google.golang.org/grpc"
)

type Server struct {
	*grpc.Server
}

func NewGRPCServer() *Server {
	return &Server{
		grpc.NewServer(
			grpc.UnaryInterceptor(unaryInterceptor()),
			grpc.StreamInterceptor(streamInterceptor()))}
}

func unaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (resp interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				log.Error("panic in GRPC unary call",
					"error", fmt.Sprintf("%v", r))

				err = fmt.Errorf("panic in GRPC unary call %s: %v: %s", info.FullMethod, r, debug.Stack())
			}
		}()
		log.Debug("GRPC unary call")
		return handler(ctx, req)
	}
}

func streamInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo,
		handler grpc.StreamHandler) (err error) {
		defer func() {
			if r := recover(); r != nil {
				log.Error("panic in GRPC stream",
					"error", fmt.Sprintf("%v", r))

				err = fmt.Errorf("panic in GRPC stream %s: %v: %s", info.FullMethod, r, debug.Stack())
			}
		}()
		log.Debug("GRPC stream call")
		return handler(srv, ss)
	}
}

func (s *Server) Start(addr string) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	go s.Server.Serve(lis) /// TODO: check error with channels

	return nil
}

func (s *Server) Stop() {
	s.Server.Stop()
}
