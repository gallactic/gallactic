package tests

import (
	"context"
	pb "github.com/gallactic/gallactic/rpc/grpc/proto3"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"testing"
)

func grpcBlockchainClient() pb.BlockChainClient {
	addr := tConfig.GRPC.ListenAddress
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	return pb.NewBlockChainClient(conn)
}

func TestGetAccounts(t *testing.T) {
	ret, err := grpcBlockchainClient().GetAccounts(context.Background(), &pb.Empty{})
	require.NoError(t, err)
	require.Equal(t, ret.Accounts[0].Account, tGenesis.Accounts()[1])
}

func TestGetValidators(t *testing.T) {
	ret, err := grpcBlockchainClient().GetValidators(context.Background(), &pb.Empty{})
	require.NoError(t, err)
	require.Equal(t, ret.Validators[0].Validator, tGenesis.Validators()[0])
}
