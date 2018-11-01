package grpc

import (
	"context"
	"fmt"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/logging/structure"
	"google.golang.org/grpc"
	"runtime/debug"
)

func NewGRPCServer(logger *logging.Logger) *grpc.Server {
	return grpc.NewServer(grpc.UnaryInterceptor(unaryInterceptor(logger)),
		grpc.StreamInterceptor(streamInterceptor(logger.WithScope("NewGRPCServer"))))
}

func unaryInterceptor(logger *logging.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (resp interface{}, err error) {
		logger = logger.With("method", info.FullMethod)
		defer func() {
			if r := recover(); r != nil {
				logger.InfoMsg("panic in GRPC unary call", structure.ErrorKey, fmt.Sprintf("%v", r))
				err = fmt.Errorf("panic in GRPC unary call %s: %v: %s", info.FullMethod, r, debug.Stack())
			}
		}()
		logger.TraceMsg("GRPC unary call")
		return handler(ctx, req)
	}
}

func streamInterceptor(logger *logging.Logger) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo,
		handler grpc.StreamHandler) (err error) {
		logger = logger.With("method", info.FullMethod,
			"is_client_stream", info.IsClientStream,
			"is_server_stream", info.IsServerStream)
		defer func() {
			if r := recover(); r != nil {
				logger.InfoMsg("panic in GRPC stream", structure.ErrorKey, fmt.Sprintf("%v", r))
				err = fmt.Errorf("panic in GRPC stream %s: %v: %s", info.FullMethod, r, debug.Stack())
			}
		}()
		logger.TraceMsg("GRPC stream call")
		return handler(srv, ss)
	}
}
