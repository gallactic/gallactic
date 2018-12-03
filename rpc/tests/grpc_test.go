package tests

import (
	"context"
	"fmt"
	"testing"

	pb "github.com/gallactic/gallactic/rpc/grpc/proto3"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

func grpcBlockchainClient() pb.BlockChainClient {
	addr := tConfig.GRPC.ListenAddress
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	return pb.NewBlockChainClient(conn)
}

func grpcTransactionClient() pb.TransactionClient {
	addr := tConfig.GRPC.ListenAddress
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	return pb.NewTransactionClient(conn)
}

func TestGetAccount(t *testing.T) {
	addr := tGenesis.Accounts()[0].Address()
	ret, err := grpcBlockchainClient().GetAccount(context.Background(), &pb.AddressRequest{Address: addr})
	require.NoError(t, err)
	require.Equal(t, ret.Account, tGenesis.Accounts()[0])
}
func TestGetValidator(t *testing.T) {
	valaddr := tGenesis.Validators()[0].Address()
	ret, err := grpcBlockchainClient().GetValidator(context.Background(), &pb.AddressRequest{Address: valaddr})
	require.NoError(t, err)
	require.Equal(t, ret.Validator, tGenesis.Validators()[0])
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

func TestGetBlock(t *testing.T) {
	ret, err := grpcBlockchainClient().GetBlock(context.Background(), &pb.BlockRequest{Height: 20})
	require.NoError(t, err)
	fmt.Println("BlockDetials", ret)
}

func TestGetBlocks(t *testing.T) {
	ret, err := grpcBlockchainClient().GetBlocks(context.Background(), &pb.BlocksRequest{MinHeight: 20, MaxHeight: 40})
	require.NoError(t, err)
	fmt.Println("BlocksDetials", ret)
}

func TestGetChainID(t *testing.T) {
	ret, err := grpcBlockchainClient().GetChainID(context.Background(), &pb.Empty{})
	require.NoError(t, err)
	fmt.Println("GetChainID", ret)
}

func TestGetLatestBlock(t *testing.T) {
	ret, err := grpcBlockchainClient().GetLatestBlock(context.Background(), &pb.BlockRequest{Height: 1000})
	require.NoError(t, err)
	fmt.Println("GetLatestBlock", ret)
}

func TestGetConsensusState(t *testing.T) {
	ret, err := grpcBlockchainClient().GetConsensusState(context.Background(), &pb.Empty{})
	require.NoError(t, err)
	fmt.Println("GetConsensusState", ret)
}
