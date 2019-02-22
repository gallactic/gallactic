package grpc

import (
	"context"
	"encoding/json"

	"github.com/gallactic/gallactic/core/consensus/tendermint/query"
	"github.com/gallactic/gallactic/core/execution"
	pb "github.com/gallactic/gallactic/rpc/grpc/proto3"
	"github.com/gallactic/gallactic/txs"
)

type transcatorService struct {
	ctx        context.Context
	nodeview   *query.NodeView
	transactor *execution.Transactor
}

var _ pb.TransactionServer = &transcatorService{}

func NewTransactorService(con context.Context, transaction *execution.Transactor, nview *query.NodeView) *transcatorService {
	return &transcatorService{
		transactor: transaction,
		nodeview:   nview,
		ctx:        con,
	}
}

// Transcation Service
func (tx *transcatorService) BroadcastTxSync(ctx context.Context, txReq *pb.TransactRequest) (*pb.ReceiptResponse, error) {
	env := new(txs.Envelope)
	err := env.UnmarshalJSON([]byte(txReq.Envelope))
	if err != nil {
		return nil, err
	}

	receipt, err := tx.transactor.BroadcastTxSync(env)
	if err != nil {
		return nil, err
	}

	rb, err := json.Marshal(receipt)
	if err != nil {
		return nil, err
	}

	return &pb.ReceiptResponse{
		Receipt: string(rb),
	}, nil
}

// Get the list of unconfirmed transaction
func (tx *transcatorService) GetUnconfirmedTxs(ctx context.Context, unconfirmreq *pb.Empty2) (*pb.UnconfirmTxsResponse, error) {
	envs, err := tx.nodeview.MempoolTransactions(-1)
	if err != nil {
		return nil, err
	}

	wrappedTxs := make([]string, len(envs))
	for i, env := range envs {
		eb, _ := json.Marshal(env)
		wrappedTxs[i] = string(eb)
	}

	return &pb.UnconfirmTxsResponse{
		Count:     int32(len(envs)),
		Envelopes: wrappedTxs,
	}, nil
}

func (tx *transcatorService) BroadcastTxAsync(ctx context.Context, txReq *pb.TransactRequest) (*pb.ReceiptResponse, error) {
	env := new(txs.Envelope)
	err := env.UnmarshalJSON([]byte(txReq.Envelope))
	receipt, err := tx.transactor.BroadcastTxAsync(env)
	if err != nil {
		return nil, err
	}

	rb, err := json.Marshal(receipt)
	if err != nil {
		return nil, err
	}

	return &pb.ReceiptResponse{
		Receipt: string(rb),
	}, nil
}
