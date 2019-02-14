package grpc

import (
	"context"
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

//Transcation Service
func (tx *transcatorService) BroadcastTx(ctx context.Context, txReq *pb.TransactRequest) (*pb.ReceiptResponse, error) {
	receipt, err := tx.transactor.BroadcastTxSync(txReq.TxEnvelope)
	if err != nil {
		return nil, err
	}

	return &pb.ReceiptResponse{
		TxReceipt: receipt,
	}, nil
}

//Get the list of unconfirmed transaction
func (tx *transcatorService) GetUnconfirmedTxs(ctx context.Context, unconfirmreq *pb.Empty2) (*pb.UnconfirmTxsResponse, error) {
	transactions, err := tx.nodeview.MempoolTransactions(-1)
	if err != nil {
		return nil, err
	}

	wrappedTxs := make([]txs.Envelope, len(transactions))
	for i, tx := range transactions {
		wrappedTxs[i] = *tx
	}

	return &pb.UnconfirmTxsResponse{
		Count:       int32(len(transactions)),
		TxEnvelopes: wrappedTxs,
	}, nil
}

