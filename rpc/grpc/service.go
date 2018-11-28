package grpc

import (
	"context"

	"github.com/gallactic/gallactic/common/binary"
	"github.com/gallactic/gallactic/core/account"
	"github.com/gallactic/gallactic/core/blockchain"
	"github.com/gallactic/gallactic/core/consensus/tendermint/query"
	"github.com/gallactic/gallactic/core/execution"
	"github.com/gallactic/gallactic/core/state"
	"github.com/gallactic/gallactic/core/validator"
	pb "github.com/gallactic/gallactic/rpc/grpc/proto3"
	"github.com/gallactic/gallactic/txs"
	"github.com/gallactic/gallactic/version"
	consensusTypes "github.com/tendermint/tendermint/consensus/types"
	tmTypes "github.com/tendermint/tendermint/types"
)

// MaxBlockLookback constant
const MaxBlockLookback = 1000

type blockchainServer struct {
	nodeview   *query.NodeView
	blockchain *blockchain.Blockchain
	state      *state.State
}

type transcatorServer struct {
	ctx        *context.Context
	nodeview   *query.NodeView
	transactor *execution.Transactor
}

type networkServer struct {
	nodeview   *query.NodeView
	blockchain *blockchain.Blockchain
}

var _ pb.TransactionServer = &transcatorServer{}
var _ pb.BlockChainServer = &blockchainServer{}
var _ pb.NetworkServer = &networkServer{}

func (s *blockchainServer) State() *state.State {
	return s.state
}
func BlockchainService(blockchain *blockchain.Blockchain, nview *query.NodeView) *blockchainServer {
	return &blockchainServer{
		blockchain: blockchain,
		nodeview:   nview,
		state:      blockchain.State(),
	}
}
func TransactorService(transction *execution.Transactor) *transcatorServer {
	return &transcatorServer{
		transactor: transction,
	}
}
func NetowrkService(blockchain *blockchain.Blockchain, nView *query.NodeView) *networkServer {
	return &networkServer{
		blockchain: blockchain,
		nodeview:   nView,
	}
}

// Blockchain Service
func (as *blockchainServer) GetAccount(ctx context.Context, param *pb.AddressRequest) (*pb.AccountResponse, error) {
	acc, err := as.state.GetAccount(param.Address)
	if err != nil {
		return nil, err
	}
	return &pb.AccountResponse{Account: acc}, nil
}

func (as *blockchainServer) GetAccounts(ctx context.Context, in *pb.Empty) (*pb.AccountsResponse, error) {
	accounts := make([]*pb.AccountResponse, 0)
	as.state.IterateAccounts(func(acc *account.Account) (stop bool) {
		if acc != nil {
			accounts = append(accounts, &pb.AccountResponse{Account: acc})
		}
		return
	})
	return &pb.AccountsResponse{
		BlockHeight: as.blockchain.LastBlockHeight(),
		Accounts:    accounts,
	}, nil
}

func (vs *blockchainServer) GetValidator(ctx context.Context, param *pb.AddressRequest) (*pb.ValidatorResponse, error) {
	val, err := vs.state.GetValidator(param.Address)
	if err != nil {
		return nil, err
	}
	return &pb.ValidatorResponse{Validator: val}, nil
}

func (vs *blockchainServer) GetValidators(context.Context, *pb.Empty) (*pb.ValidatorsResponse, error) {
	validators := make([]*pb.ValidatorResponse, 0)
	vs.state.IterateValidators(func(val *validator.Validator) (stop bool) {
		if val != nil {
			validators = append(validators, &pb.ValidatorResponse{Validator: val})
		}
		return true
	})
	return &pb.ValidatorsResponse{
		Validators:  validators,
		BlockHeight: vs.blockchain.LastBlockHeight(),
	}, nil
}
func (s *blockchainServer) GetStorage(ctx context.Context, storage *pb.StorageAtRequest) (*pb.StorageResponse, error) {
	value, err := s.state.GetStorage(storage.Address, binary.LeftPadWord256(storage.Key))
	if err != nil {
		return nil, err
	}
	if value == binary.Zero256 {
		return &pb.StorageResponse{Key: storage.Key, Value: nil}, nil
	}
	return &pb.StorageResponse{Key: storage.Key, Value: value.UnpadLeft()}, nil
}

func (s *blockchainServer) GetStorageAt(ctx context.Context, storage *pb.StorageAtRequest) (*pb.StorageResponse, error) {
	value, err := s.state.GetStorage(storage.Address, binary.LeftPadWord256(storage.Key))
	if err != nil {
		return nil, err
	}
	if value == binary.Zero256 {
		return &pb.StorageResponse{Key: storage.Key, Value: nil}, nil
	}
	return &pb.StorageResponse{Key: storage.Key, Value: value.UnpadLeft()}, nil
}

func (s *blockchainServer) Getstatus(ctx context.Context, in *pb.Empty) (*pb.StatusResponse, error) {
	latestHeight := s.blockchain.LastBlockHeight()
	var latestBlockMeta *tmTypes.BlockMeta
	var latestBlockHash []byte
	var latestBlockTime int64
	if latestHeight != 0 {
		latestBlockMeta = s.nodeview.BlockStore().LoadBlockMeta(int64(latestHeight))
		latestBlockHash = latestBlockMeta.Header.Hash()
		latestBlockTime = latestBlockMeta.Header.Time.UnixNano()
	}
	publicKey, err := s.nodeview.PrivValidatorPublicKey()
	if err != nil {
		return nil, err
	}
	return &pb.StatusResponse{
		//NodeInfo:          s.nodeview.deo,
		GenesisHash:       s.blockchain.GenesisHash(),
		PubKey:            publicKey,
		LatestBlockHash:   latestBlockHash,
		LatestBlockHeight: latestHeight,
		LatestBlockTime:   latestBlockTime,
		NodeVersion:       version.Version,
	}, err
}

func (s *blockchainServer) GetBlock(ctx context.Context, block *pb.BlockRequest) (*pb.BlockResponse, error) {
	Block := s.nodeview.BlockStore().LoadBlock(int64(block.Height))
	Blockmeta := s.nodeview.BlockStore().LoadBlockMeta(int64(block.Height))
	return &pb.BlockResponse{
		Block:     Block,
		BlockMeta: Blockmeta,
	}, nil

}

func (s *blockchainServer) GetBlocks(ctx context.Context, blocks *pb.BlocksRequest) (*pb.BlocksResponse, error) {
	latestHeight := s.blockchain.LastBlockHeight()
	if blocks.MinHeight == 0 {
		blocks.MinHeight = 1
	}
	if blocks.MaxHeight == 0 || latestHeight < blocks.MaxHeight {
		blocks.MaxHeight = latestHeight
	}
	if blocks.MaxHeight > blocks.MinHeight && blocks.MaxHeight-blocks.MinHeight > MaxBlockLookback {
		blocks.MinHeight = blocks.MaxHeight - MaxBlockLookback
	}
	var blockMetas []tmTypes.BlockMeta
	for height := blocks.MaxHeight; height >= blocks.MinHeight; height-- {
		blockMeta := s.nodeview.BlockStore().LoadBlockMeta(int64(height))
		blockMetas = append(blockMetas, *blockMeta)
	}

	return &pb.BlocksResponse{
		LastHeight: latestHeight,
		BlockMeta:  blockMetas,
	}, nil

}

func (s *blockchainServer) GetGenesis(context.Context, *pb.Empty) (*pb.GenesisResponse, error) {
	gen := s.blockchain.Genesis()
	return &pb.GenesisResponse{
		Genesis: gen,
	}, nil
}

func (s *blockchainServer) GetChainID(context.Context, *pb.Empty) (*pb.ChainResponse, error) {
	return &pb.ChainResponse{
		ChainName:   s.blockchain.Genesis().ChainName(),
		ChainId:     s.blockchain.ChainID(),
		GenesisHash: s.blockchain.GenesisHash(),
	}, nil

}

func (s *blockchainServer) GetLatestBlock(context.Context, *pb.BlockRequest) (*pb.BlockResponse, error) {
	latestHeight := s.blockchain.LastBlockHeight()
	block := s.nodeview.BlockStore().LoadBlock(int64(latestHeight))
	blockMeta := s.nodeview.BlockStore().LoadBlockMeta(int64(latestHeight))
	return &pb.BlockResponse{
		BlockMeta: blockMeta,
		Block:     block,
	}, nil
}

func (s *blockchainServer) GetConsensusState(context.Context, *pb.Empty) (*pb.ConsensusResponse, error) {
	peerRound := make([]consensusTypes.PeerRoundState, 0)
	peerRoundState, err := s.nodeview.PeerRoundStates()
	for _, pr := range peerRoundState {
		peerRound = append(peerRound, *pr)
	}
	if err != nil {
		return nil, err
	}
	return &pb.ConsensusResponse{
		RoundState:      s.nodeview.RoundState().RoundStateSimple(),
		PeerRoundStates: peerRound,
	}, nil

}

func (s *blockchainServer) GetBlockTxs(ctx context.Context, block *pb.BlockRequest) (*pb.BlockTxsResponse, error) {
	result, err := s.GetBlock(ctx, block)
	if err != nil {
		return nil, err
	}
	txsBuff := result.Block.Txs
	txList := make([]txs.Envelope, len(txsBuff))
	for i, txBuff := range txsBuff {
		tx, err := txs.NewAminoCodec().DecodeTx(txBuff)
		if err != nil {
			return nil, err
		}
		txList[i] = *tx
	}
	return &pb.BlockTxsResponse{
		Count: int32(len(txsBuff)),
		Txs:   txList,
	}, nil

}

//Network service
func (s *networkServer) GetNetworkInfo(context.Context, *pb.Empty1) (*pb.NetInfoResponse, error) {
	//listening := s.blockchain.IsListening()

	var contexts context.Context
	//var listeners []string
	// for _, listener := range s.nodeview.Listeners() {
	// 	listeners = append(listeners, listener.String())
	// }
	peers, err := s.GetPeers(contexts, nil)
	if err != nil {
		return nil, err
	}
	return &pb.NetInfoResponse{
		// Listening: listening,
		// Listeners: listeners,
		Peers: peers.Peer,
	}, nil

	return nil, nil
}

func (ns *networkServer) GetPeers(context.Context, *pb.Empty1) (*pb.PeerResponse, error) {
	peers := make([]*pb.Peer, ns.nodeview.Peers().Size())
	for i, peer := range ns.nodeview.Peers().List() {
		peers[i] = &pb.Peer{
			//NodeInfo:   peer.NodeInfo(),
			IsOutbound: peer.IsOutbound(),
		}
	}

	return &pb.PeerResponse{
		Peer: peers,
	}, nil
}

//Transcation Service
func (tx *transcatorServer) BroadcastTx(ctx context.Context, txreq *pb.TransactRequest) (*pb.ReceiptResponse, error) {
	txhash, err := tx.transactor.BroadcastTx(txreq.Txs)
	if err != nil {
		return nil, err
	}

	return &pb.ReceiptResponse{
		TxHash: txhash,
	}, nil
}

func (tx *transcatorServer) GetUnconfirmedTxs(ctx context.Context, unconfirmreq *pb.UnconfirmedTxsRequest) (*pb.UnconfirmTxsResponse, error) {
	transactions, err := tx.nodeview.MempoolTransactions(int(unconfirmreq.MaxTxs))
	if err != nil {
		return nil, err
	}

	wrappedTxs := make([]txs.Envelope, len(transactions))
	for i, tx := range transactions {
		wrappedTxs[i] = *tx
	}

	return &pb.UnconfirmTxsResponse{
		Count: int32(len(transactions)),
		Txs:   wrappedTxs,
	}, nil
}
