package grpc

import (
	"context"
	"github.com/gallactic/gallactic/common/binary"
	"github.com/gallactic/gallactic/core/account"
	"github.com/gallactic/gallactic/core/blockchain"
	"github.com/gallactic/gallactic/core/consensus/tendermint/p2p"
	"github.com/gallactic/gallactic/core/consensus/tendermint/query"
	"github.com/gallactic/gallactic/core/execution"
	"github.com/gallactic/gallactic/core/state"
	"github.com/gallactic/gallactic/core/validator"
	"github.com/gallactic/gallactic/crypto"
	pb "github.com/gallactic/gallactic/rpc/grpc/proto3"
	"github.com/gallactic/gallactic/txs"
	"github.com/gallactic/gallactic/version"
	consensusTypes "github.com/tendermint/tendermint/consensus/types"
    net "github.com/tendermint/tendermint/p2p"
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
	ctx        context.Context
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
func TransactorService(con context.Context, transaction *execution.Transactor, nview *query.NodeView) *transcatorServer {
	return &transcatorServer{
		transactor: transaction,
		nodeview:   nview,
		ctx:        con,
	}
}
func NetworkService(blockchain *blockchain.Blockchain, nView *query.NodeView) *networkServer {
	return &networkServer{
		blockchain: blockchain,
		nodeview:   nView,
	}
}

// Blockchain Service
func (as *blockchainServer) GetAccount(ctx context.Context, param *pb.AddressRequest) (*pb.AccountResponse, error) {
	addr, err := crypto.AddressFromString(param.Address)
	if err != nil {
		return nil, err
	}
	acc, err := as.state.GetAccount(addr)
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
		return false
	})
	return &pb.AccountsResponse{
		BlockHeight: as.blockchain.LastBlockHeight(),
		Accounts:    accounts,
	}, nil
}

func (vs *blockchainServer) GetValidator(ctx context.Context, param *pb.AddressRequest) (*pb.ValidatorResponse, error) {
	addr, err := crypto.AddressFromString(param.Address)
	if err != nil {
		return nil, err
	}
	val, err := vs.state.GetValidator(addr)
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
		return false
	})
	return &pb.ValidatorsResponse{
		Validators:  validators,
		BlockHeight: vs.blockchain.LastBlockHeight(),
	}, nil
}

func (s *blockchainServer) GetStorage(ctx context.Context, storage *pb.StorageRequest) (*pb.StorageResponse, error) {
	var storageItems []pb.StorageItem

	storageaddr, err := crypto.AddressFromString(storage.Address)
	if err != nil {
		return nil, err
	}

	s.state.IterateStorage(storageaddr, func(key, value binary.Word256) (stop bool) {
		storageItems = append(storageItems, pb.StorageItem{Key: key.UnpadLeft(), Value: value.UnpadLeft()})
<<<<<<< refs/remotes/gallactic/develop
		return false
=======
		return
>>>>>>> changes in get_stroage method
	})
	return &pb.StorageResponse{
		StorageItems: storageItems,
	}, nil

}

func (s *blockchainServer) GetStorageAt(ctx context.Context, storage *pb.StorageAtRequest) (*pb.StorageAtResponse, error) {
	storageaddr, err := crypto.AddressFromString(storage.Address)
	if err != nil {
		return nil, err
	}
	value, err := s.state.GetStorage(storageaddr, binary.LeftPadWord256(storage.Key))
	if err != nil {
		return nil, err
	}
	if value == binary.Zero256 {
		return &pb.StorageAtResponse{Key: storage.Key, Value: nil}, nil
	}
	return &pb.StorageAtResponse{Key: storage.Key, Value: value.UnpadLeft()}, nil
}

func (s *blockchainServer) GetStatus(ctx context.Context, in *pb.Empty) (*pb.StatusResponse, error) {
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
		//NodeInfo:          s.nodeview.NodeInfo(),
		GenesisHash:       s.blockchain.GenesisHash(),
		PubKey:            publicKey,
		LatestBlockHash:   latestBlockHash,
		LatestBlockHeight: latestHeight,
		LatestBlockTime:   latestBlockTime,
		NodeVersion:       version.Version,
	}, err
}

func (s *blockchainServer) GetBlock(ctx context.Context, block *pb.BlockRequest) (*pb.BlockResponse, error) {
	height := int64(block.Height)
	if height == 0 {
		height = s.nodeview.BlockStore().Height()
	}
	Block := s.nodeview.BlockStore().LoadBlock(height)
	Blockmeta := s.nodeview.BlockStore().LoadBlockMeta(height)
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

func (s *blockchainServer) GetLatestBlock(context.Context, *pb.Empty) (*pb.BlockResponse, error) {
	latestHeight := s.blockchain.LastBlockHeight()
	block := s.nodeview.BlockStore().LoadBlock(int64(latestHeight))
	blockMeta := s.nodeview.BlockStore().LoadBlockMeta(int64(latestHeight))
	return &pb.BlockResponse{
		BlockMeta: blockMeta,
		Block:     block,
	}, nil
}

func (s *blockchainServer) GetBlockchainInfo(ctx context.Context, blockinfo *pb.Empty) (*pb.BlockchainInfoResponse, error) {
	res := &pb.BlockchainInfoResponse{
		LastBlockHeight: s.blockchain.LastBlockHeight(),
		LastBlockHash:   s.blockchain.LastBlockHash(),
		LastBlockTime:   s.blockchain.LastBlockTime(),
	}
	return res, nil
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
		txEnv := new(txs.Envelope)

		if err := txEnv.Decode(txBuff); err != nil {
			return nil, err
		}
		txList[i] = *txEnv
	}
	return &pb.BlockTxsResponse{
		Count: int32(len(txsBuff)),
		Txs:   txList,
	}, nil

}

//Network service
func (s *networkServer) GetNetworkInfo(context.Context, *pb.Empty1) (*pb.NetInfoResponse, error) {
	var contexts context.Context
	peers, err := s.GetPeers(contexts, nil)
	if err != nil {
		return nil, err
	}
	return &pb.NetInfoResponse{
		Peers: peers.Peer,
	}, nil

}

func (ns *networkServer) GetPeers(context.Context, *pb.Empty1) (*pb.PeerResponse, error) {
	peers := make([]*pb.Peer, ns.nodeview.Peers().Size())
	for i, peer := range ns.nodeview.Peers().List() {
		ni := new(p2p.GNodeInfo)
		tmni, _ := peer.NodeInfo().(*net.DefaultNodeInfo)
		ni.ID_ = tmni.ID_
		ni.Network = tmni.Network
		ni.ProtocolVersion = tmni.ProtocolVersion
		ni.Version = tmni.Version
		ni.Channels = tmni.Channels
		ni.ListenAddr = tmni.ListenAddr
		ni.Moniker = tmni.Moniker
		peers[i] = &pb.Peer{
			NodeInfo:   *ni,
			IsOutbound: peer.IsOutbound(),
		}
	}
	return &pb.PeerResponse{
		Peer: peers,
	}, nil
}

//Transcation Service
func (tx *transcatorServer) BroadcastTx(ctx context.Context, txReq *pb.TransactRequest) (*pb.ReceiptResponse, error) {
	receipt, err := tx.transactor.BroadcastTx(txReq.TxEnvelope)
	if err != nil {
		return nil, err
	}

	return &pb.ReceiptResponse{
		TxReceipt: receipt,
	}, nil
}

func (tx *transcatorServer) GetUnconfirmedTxs(ctx context.Context, unconfirmreq *pb.Empty2) (*pb.UnconfirmTxsResponse, error) {
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



