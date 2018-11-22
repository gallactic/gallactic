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

var _ TransactionServer = &transcatorServer{}
var _ BlockChainServer = &blockchainServer{}
var _ NetworkServer = &networkServer{}

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
func (as *blockchainServer) GetAccount(ctx context.Context, param *AddressRequest) (*AccountResponse, error) {
	acc, err := as.state.GetAccount(param.Address)
	if err != nil {
		return nil, err
	}
	return &AccountResponse{Account: acc}, nil

}

func (as *blockchainServer) GetAccounts(ctx context.Context, in *Empty) (*AccountsResponse, error) {
	accounts := make([]*Account, 0)
	as.state.IterateAccounts(func(acc *account.Account) (stop bool) {
		if acc != nil {
			accounts = append(accounts, &Account{Account: acc})
		}
		return
	})
	return &AccountsResponse{
		BlockHeight: as.blockchain.LastBlockHeight(),
		Account:     accounts,
	}, nil
}
func (vs *blockchainServer) GetValidator(ctx context.Context, param *AddressRequest) (*ValidatorResponse, error) {
	val, err := vs.state.GetValidator(param.Address)
	if err != nil {
		return nil, err
	}
	return &ValidatorResponse{Validator: val}, nil

}

func (vs *blockchainServer) GetValidators(context.Context, *Empty) (*ValidatorsResponse, error) {
	validators := make([]validator.Validator, 0)
	vs.state.IterateValidators(func(val *validator.Validator) (stop bool) {
		validators = append(validators, *val)
		return true
	})
	return &ValidatorsResponse{
		Validators:  validators,
		BlockHeight: vs.blockchain.LastBlockHeight(),
	}, nil

}
func (s *blockchainServer) GetStorage(ctx context.Context, storage *StorageAtRequest) (*StorageResponse, error) {
	value, err := s.state.GetStorage(storage.Address, binary.LeftPadWord256(storage.Key))
	if err != nil {
		return nil, err
	}
	if value == binary.Zero256 {
		return &StorageResponse{Key: storage.Key, Value: nil}, nil
	}
	return &StorageResponse{Key: storage.Key, Value: value.UnpadLeft()}, nil
}
func (s *blockchainServer) GetStorageAt(ctx context.Context, storage *StorageAtRequest) (*StorageResponse, error) {
	value, err := s.state.GetStorage(storage.Address, binary.LeftPadWord256(storage.Key))
	if err != nil {
		return nil, err
	}
	if value == binary.Zero256 {
		return &StorageResponse{Key: storage.Key, Value: nil}, nil
	}
	return &StorageResponse{Key: storage.Key, Value: value.UnpadLeft()}, nil
}

func (s *blockchainServer) Getstatus(ctx context.Context, in *Empty) (*StatusResponse, error) {
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
	return &StatusResponse{
		NodeInfo:          s.nodeview.NodeInfo(),
		GenesisHash:       s.blockchain.GenesisHash(),
		PubKey:            publicKey,
		LatestBlockHash:   latestBlockHash,
		LatestBlockHeight: latestHeight,
		LatestBlockTime:   latestBlockTime,
		NodeVersion:       version.Version,
	}, err

}

func (s *blockchainServer) GetBlock(ctx context.Context, block *BlockRequest) (*BlockResponse, error) {

	// //TODO changes to be made in vendor/tendermint for block and blockmeta.
	// Block := s.nodeview.BlockStore().LoadBlock(int64(block.Height))
	// Blockmeta := s.nodeview.BlockStore().LoadBlockMeta(int64(block.Height))
	// return &BlockResponse{
	// 	Block:     Block,
	// 	BlockMeta: Blockmeta,
	// }, nil
	return nil, nil
}

func (s *blockchainServer) GetBlocks(ctx context.Context, blocks *BlocksRequest) (*BlocksResponse, error) {

	//latestHeight := s.blockchain.LastBlockHeight()
	// if blocks.MinHeight == 0 {
	// 	blocks.MinHeight = 1
	// }
	// if blocks.MaxHeight == 0 || latestHeight < blocks.MaxHeight {
	// 	blocks.MaxHeight = latestHeight
	// }
	// if blocks.MaxHeight > blocks.MinHeight && blocks.MaxHeight-blocks.MinHeight > MaxBlockLookback {
	// 	blocks.MinHeight = blocks.MaxHeight - MaxBlockLookback
	// }

	//TODO changes to be made in vendor/tendermint  blockmeta.
	// var blockMetas []tmTypes.BlockMeta
	// for height := blocks.MaxHeight; height >= blocks.MinHeight; height-- {
	// 	blockMeta := s.nodeview.BlockStore().LoadBlockMeta(int64(height))
	// 	blockMetas = append(blockMetas, *blockMeta)
	// }

	// return &BlocksResponse{
	// 	LastHeight: latestHeight,
	// 	BlockMeta:  blockMetas,
	// }, nil
	return nil, nil
}

func (s *blockchainServer) GetGenesis(context.Context, *Empty) (*GenesisResponse, error) {
	gen := s.blockchain.Genesis()
	return &GenesisResponse{
		Genesis: gen,
	}, nil
}

func (s *blockchainServer) GetChainID(context.Context, *Empty) (*ChainResponse, error) {
	return &ChainResponse{
		ChainName:   s.blockchain.Genesis().ChainName(),
		ChainId:     s.blockchain.ChainID(),
		GenesisHash: s.blockchain.GenesisHash(),
	}, nil

}

func (s *blockchainServer) GetLatestBlock(context.Context, *BlockRequest) (*BlockResponse, error) {
	latestHeight := s.blockchain.LastBlockHeight()

	//TODO changes to be made in vendor/tendermint  blockmeta.
	block := s.nodeview.BlockStore().LoadBlock(int64(latestHeight))
	blockMeta := s.nodeview.BlockStore().LoadBlockMeta(int64(latestHeight))
	return &BlockResponse{
		BlockMeta: blockMeta,
		Block:     block,
	}, nil
}
func (s *blockchainServer) GetConsensusState(context.Context, *Empty) (*ConsensusResponse, error) {
	peerRound := make([]consensusTypes.PeerRoundState, 0)
	//TODO changes to be made in vendor/tendermint  for PeerRoundStates and RoundState.
	peerRoundState, err := s.nodeview.PeerRoundStates()
	for _, pr := range peerRoundState {
		peerRound = append(peerRound, *pr)
	}
	if err != nil {
		return nil, err
	}
	return &ConsensusResponse{
		RoundState:      s.nodeview.RoundState().RoundStateSimple(),
		PeerRoundStates: peerRound,
	}, nil

}

func (s *blockchainServer) GetBlockTxs(ctx context.Context, block *BlockRequest) (*BlockTxsResponse, error) {
	//TODO changes to be made in vendor/tendermint  for Block.
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

	return &BlockTxsResponse{
		Count: int32(len(txsBuff)),
		Txs:   txList,
	}, nil

}

//Network service
func (s *networkServer) GetNetworkInfo(context.Context, *Empty) (*NetInfoResponse, error) {
	// listening := s.nodeview.IsListening()
	// fmt.Println("is listening", listening)
	// var contexts context.Context
	// var listeners []string
	// for _, listener := range s.nodeview.Listeners() {
	// 	listeners = append(listeners, listener.String())
	// }
	// peers, err := s.GetPeers(contexts, nil)
	// fmt.Println("peers", peers)
	// if err != nil {
	// 	return nil, err
	// }
	// return &NetInfoResponse{
	// 	Listening: listening,
	// 	Listeners: listeners,
	// 	Peers:     peers.Peer,
	// }, nil

	return nil, nil
}

func (ns *networkServer) GetPeers(context.Context, *Empty) (*PeerResponse, error) {

	peers := make([]*Peer, ns.nodeview.Peers().Size())
	for i, peer := range ns.nodeview.Peers().List() {
		peers[i] = &Peer{
			NodeInfo:   peer.NodeInfo(),
			IsOutbound: peer.IsOutbound(),
		}
	}
	return &PeerResponse{
		Peer: peers,
	}, nil
}

//Transcation Service
func (tx *transcatorServer) BroadcastTx(ctx context.Context, txreq *TransactRequest) (*ReceiptResponse, error) {

	txhash, err := tx.transactor.BroadcastTx(txreq.Txs)
	if err != nil {
		return nil, err
	}
	return &ReceiptResponse{
		TxHash: txhash,
	}, nil
}

func (tx *transcatorServer) GetUnconfirmedTxs(ctx context.Context, unconfirmreq *UnconfirmedTxsRequest) (*UnconfirmTxsResponse, error) {
	// Get all transactions for now
	transactions, err := tx.nodeview.MempoolTransactions(int(unconfirmreq.MaxTxs))
	if err != nil {
		return nil, err
	}

	wrappedTxs := make([]txs.Envelope, len(transactions))
	for i, tx := range transactions {
		wrappedTxs[i] = *tx
	}

	return &UnconfirmTxsResponse{
		Count: int32(len(transactions)),
		Txs:   wrappedTxs,
	}, nil

}
