package grpc

import (
	"context"

	"github.com/gallactic/gallactic/core/blockchain"
	"github.com/gallactic/gallactic/rpc/grpc/proto3"
	pb "github.com/gallactic/gallactic/rpc/grpc/proto3"
)

type eventsServer struct {
}

var _ pb.EventsServer = &eventsServer{}

func EventsServer(bc *blockchain.Blockchain) *eventsServer {
	return &eventsServer{}
}

func (srv *eventsServer) Subscribe(context.Context, *proto3.SubscribeRequest) (*proto3.SubscribeResponse, error) {

	return &pb.SubscribeResponse{}, nil
}
