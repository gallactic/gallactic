package tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/gallactic/gallactic/crypto"
	"github.com/gallactic/gallactic/keystore/key"
	"github.com/gallactic/gallactic/txs"
	"github.com/gallactic/gallactic/txs/tx"

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

func TestBlockchainMethods(t *testing.T) {
	client := grpcBlockchainClient()

	//
	addr := tGenesis.Accounts()[1].Address()
	ret1, err := client.GetAccount(context.Background(), &pb.AddressRequest{Address: addr.String()})
	require.NoError(t, err)
	require.Equal(t, ret1.Account, tGenesis.Accounts()[1])

	//
	valAddr := tGenesis.Validators()[0].Address()
	ret2, err := client.GetValidator(context.Background(), &pb.AddressRequest{Address: valAddr.String()})
	require.NoError(t, err)
	require.Equal(t, ret2.Validator.Address, valAddr.String())

	//
	ret3, err := client.GetAccounts(context.Background(), &pb.Empty{})
	require.NoError(t, err)
	require.Equal(t, ret3.Accounts[0].Account, tGenesis.Accounts()[1])

	//
	ret4, err := client.GetValidators(context.Background(), &pb.Empty{})
	require.NoError(t, err)
	require.Equal(t, ret4.Validators[0].Address, valAddr.String())

	//
	ret5, err := client.GetGenesis(context.Background(), &pb.Empty{})
	require.NoError(t, err)
	require.Equal(t, ret5.Genesis, tGenesis)

	//
	ret6, err := client.GetChainID(context.Background(), &pb.Empty{})
	require.NoError(t, err)
	require.Equal(t, ret6.ChainId, tGenesis.ChainID())

	//
	ret7, err := client.GetConsensusState(context.Background(), &pb.Empty{})
	require.NoError(t, err)
	fmt.Println("GetConsensusState", ret7)

	// wait until blockchain starts...
	for {
		status, err := client.GetStatus(context.Background(), &pb.Empty{})
		require.NoError(t, err)
		if status.LatestBlockHeight > 2 {
			break
		}
		time.Sleep(100)
	}

	//
	ret8, err := client.GetLatestBlock(context.Background(), &pb.Empty{})
	require.NoError(t, err)
	fmt.Println("GetLatestBlock", ret8)

	//
	ret9, err := client.GetBlock(context.Background(), &pb.BlockRequest{Height: uint64(ret8.Block.Header.Height)})
	require.NoError(t, err)
	require.Equal(t, ret9, ret8)

	ret10, err := client.GetBlocks(context.Background(), &pb.BlocksRequest{MinHeight: 1, MaxHeight: 10})
	require.NoError(t, err)
	//require.Equal(t, ret10[0], ret8)
	fmt.Println("GetLatestBlock", ret10)
}

func TestTransactionMethods(t *testing.T) {
	client := grpcTransactionClient()

	acc_1 := tGenesis.Accounts()[1]
	acc_2 := tGenesis.Accounts()[2]
	bal_1 := acc_1.Balance()
	bal_2 := acc_2.Balance()
	k, err := key.DecryptKeyFile(tWorkingDir+"/keys/"+acc_1.Address().String()+".json", "")
	require.NoError(t, err)
	signer := crypto.NewAccountSigner(k.PrivateKey())
	tx, _ := tx.NewSendTx(acc_1.Address(), acc_2.Address(), 1, 1000, 200)
	env := txs.Enclose(tGenesis.ChainName(), tx)
	require.NoError(t, env.Sign(signer))

	bcClient := grpcBlockchainClient()

	ret1, err := client.BroadcastTx(context.Background(), &pb.TransactRequest{TxEnvelope: env})
	require.NoError(t, err)
	require.Equal(t, env.Hash(), ret1.TxReceipt.Hash.Bytes())

	// wait for new block and check balance
	for {
		retTx, err := bcClient.GetTx(context.Background(), &pb.TxRequest{TxHash: ret1.TxReceipt.Hash.String()})
		if err == nil && retTx != nil {
			break
		}
	}

	ret2, _ := bcClient.GetAccount(context.Background(), &pb.AddressRequest{Address: acc_1.Address().String()})
	ret3, _ := bcClient.GetAccount(context.Background(), &pb.AddressRequest{Address: acc_2.Address().String()})

	require.Equal(t, ret2.Account.Balance(), bal_1-1200)
	require.Equal(t, ret3.Account.Balance(), bal_2+1000)
}
