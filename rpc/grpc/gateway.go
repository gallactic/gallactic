package grpc

import (
	"context"
	"flag"
	"net/http"

	pb "github.com/gallactic/gallactic/rpc/grpc/proto3"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"
)

//GRPC GATEWAY

const (
	grpcPort = "50051"
)

var (
	swaggerDir = flag.String("swagger_dir", "template", "path to the directory which contains swagger definitions")
)

func (s *Server) StartGateway(ctx context.Context, gatewayAddr, grpcAddr string) error {

	getEndpoint := flag.String("get", gatewayAddr, "endpoint of Gallactic(GET)")

	mux := runtime.NewServeMux(runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONBuiltin{}))
	opts := []grpc.DialOption{grpc.WithInsecure()}

	if err := pb.RegisterBlockChainHandlerFromEndpoint(ctx, mux, *getEndpoint, opts); err != nil {
		return err
	}

	if err := pb.RegisterNetworkHandlerFromEndpoint(ctx, mux, *getEndpoint, opts); err != nil {
		return err
	}

	if err := pb.RegisterTransactionHandlerFromEndpoint(ctx, mux, *getEndpoint, opts); err != nil {
		return err
	}

	go http.ListenAndServe(grpcAddr, mux) /// TODO: check error with channels

	return nil
}
