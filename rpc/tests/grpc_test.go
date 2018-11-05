package tests

import (
	"context"
	"testing"

	ggrpc "github.com/gallactic/gallactic/rpc/grpc"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

func grpcAccountsClient() ggrpc.AccountsClient {
	addr := tConfig.GRPC.ListenAddress

	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	return ggrpc.NewAccountsClient(conn)
}

func TestGetAccount(t *testing.T) {
	ret, err := grpcAccountsClient().GetAccounts(context.Background(), &ggrpc.Empty{})
	require.NoError(t, err)
	require.Equal(t, ret.Account[0].Account, tGenesis.Accounts()[1])
}

func TestGetValidator(t *testing.T) {
	ret, err := grpcAccountsClient().GetValidators(context.Background(), &ggrpc.Empty{})
	require.NoError(t, err)
	require.Equal(t, ret.Validators[0], *tGenesis.Validators()[0])
}
