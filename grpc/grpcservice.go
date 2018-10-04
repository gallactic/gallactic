package grpc

import (
	"context"
	"github.com/gallactic/gallactic/common/binary"
	"github.com/gallactic/gallactic/core/blockchain"
	query "github.com/gallactic/gallactic/core/consensus/tendermint/query"
	"github.com/gallactic/gallactic/core/execution"
	"github.com/gallactic/gallactic/core/state"
	"github.com/gallactic/gallactic/core/validator"
	rpc "github.com/gallactic/gallactic/rpc"
	"github.com/gallactic/gallactic/txs"
	"github.com/gallactic/gallactic/version"
	tmTypes "github.com/tendermint/tendermint/types"
)

// MaxBlockLookback constant
const MaxBlockLookback = 1000

// Service struct
type accountServer struct {
	accounts   *state.State
	blockchain *blockchain.Blockchain
}

type blockchainServer struct {
	blockchain *blockchain.Blockchain
	nodeview   *query.NodeView
}

type networkServer struct {
	ctx      *context.Context
	nodeview *query.NodeView
}

type TranscatorServer struct {
	ctx      *context.Context
	nodeview *query.NodeView
	transact *execution.Transactor
}

func AccountService(State *state.State) *accountServer {
	var blockchain *blockchain.Blockchain
	return &accountServer{
		accounts:   State,
		blockchain: blockchain,
	}
}

func BlockchainService(blockchain *blockchain.Blockchain) *blockchainServer {
	var query *query.NodeView
	return &blockchainServer{
		blockchain: blockchain,
		nodeview:   query,
	}
}

func NetowrkService(query *query.NodeView) *networkServer {
	var contexts *context.Context
	return &networkServer{
		nodeview: query,
		ctx:      contexts,
	}
}

func TransactorService(transactor *execution.Transactor) *TranscatorServer {
	return &TranscatorServer{
		transact: transactor,
	}
}

func (as *accountServer) GetValidators(context.Context, *Empty) (*ValidatorOutput, error) {
	validators := make([]*validator.Validator, 0)
	as.blockchain.State().IterateValidators(func(val *validator.Validator) (stop bool) {
		validators = append(validators, val)
		return
	})
	return &ValidatorOutput{
		BlockHeight:         as.blockchain.LastBlockHeight(),
		BondedValidators:    validators,
		UnbondingValidators: nil,
	}, nil
}

var _ NetworkServer = &networkServer{}
var _ AccountsServer = &accountServer{}
var _ BlockChainServer = &blockchainServer{}
var _ TransactionServer = &TranscatorServer{}

// Account Service
func (s *accountServer) GetAccount(ctx context.Context, param *AddressParam) (*AccountOutput, error) {
	acc, err := s.accounts.GetAccount(param.Address)
	if err != nil {
		return nil, err
	}
	return &AccountOutput{Account: *acc}, nil
}

func (s *accountServer) GetStorage(ctx context.Context, storage *StorageAtInput) (*StorageOutput, error) {
	value, err := s.accounts.GetStorage(storage.Address, binary.LeftPadWord256(storage.Key))
	if err != nil {
		return nil, err
	}
	if value == binary.Zero256 {
		return &StorageOutput{Key: storage.Key, Value: nil}, nil
	}
	return &StorageOutput{Key: storage.Key, Value: value.UnpadLeft()}, nil
}

func (s *accountServer) GetStorageAt(ctx context.Context, storage *StorageAtInput) (*StorageOutput, error) {
	value, err := s.accounts.GetStorage(storage.Address, binary.LeftPadWord256(storage.Key))
	if err != nil {
		return nil, err
	}
	if value == binary.Zero256 {
		return &StorageOutput{Key: storage.Key, Value: nil}, nil
	}
	return &StorageOutput{Key: storage.Key, Value: value.UnpadLeft()}, nil
}

func (s *accountServer) GetAccounts(ctx context.Context, filterinput *AccountParam) (*AccountsOutput, error) {
	//input := filterinput.Account
	//  list, err := s.accounts.IterateAccounts(func(account *account.Account) bool {
	// 	return filter.Match(account)
	//  })
	// 	if err != nil {
	// 		return nil, RPCErrorInternalError, err
	//  	}
	return &AccountsOutput{}, nil

}

//Blockchain Service

func (s *blockchainServer) GetBlock(ctx context.Context, block *BlockInput) (*BlockOutput, error) {
	return &BlockOutput{
		Block:     s.nodeview.BlockStore().LoadBlock(int64(block.Height)),
		BlockMeta: s.nodeview.BlockStore().LoadBlockMeta(int64(block.Height)),
	}, nil
}

func (s *blockchainServer) GetBlocks(ctx context.Context, in *BlocksInput) (*BlocksOutput, error) {

	latestHeight := s.blockchain.LastBlockHeight()
	if in.MinHeight == 0 {
		in.MinHeight = 1
	}
	if in.MaxHeight == 0 || latestHeight < in.MaxHeight {
		in.MaxHeight = latestHeight
	}
	if in.MaxHeight > in.MinHeight && in.MaxHeight-in.MinHeight > MaxBlockLookback {
		in.MinHeight = in.MaxHeight - MaxBlockLookback
	}

	var blockMetas []*tmTypes.BlockMeta
	for height := in.MaxHeight; height >= in.MinHeight; height-- {
		blockMeta := s.nodeview.BlockStore().LoadBlockMeta(int64(height))
		blockMetas = append(blockMetas, blockMeta)
	}

	return &BlocksOutput{
		LastHeight: int(latestHeight),
		BlockMeta:  blockMetas,
	}, nil

}

func (s *blockchainServer) Getstatus(ctx context.Context, in *Empty) (*StatusOutput, error) {

	latestHeight := s.blockchain.LastBlockHeight()
	var (
		latestBlockMeta *tmTypes.BlockMeta
		latestBlockHash []byte
		latestBlockTime int64
	)
	if latestHeight != 0 {
		latestBlockMeta = s.nodeview.BlockStore().LoadBlockMeta(int64(latestHeight))
		latestBlockHash = latestBlockMeta.Header.Hash()
		latestBlockTime = latestBlockMeta.Header.Time.UnixNano()
	}
	publicKey, err := s.nodeview.PrivValidatorPublicKey()
	if err != nil {
		return nil, err
	}
	return &StatusOutput{
		NodeInfo:          s.nodeview.NodeInfo(),
		GenesisHash:       s.blockchain.GenesisHash(),
		PubKey:            publicKey,
		LatestBlockHash:   latestBlockHash,
		LatestBlockHeight: latestHeight,
		LatestBlockTime:   latestBlockTime,
		NodeVersion:       version.Version,
	}, nil

}

func (s *blockchainServer) GetLatestBlock(ctx context.Context, in *BlockInput) (*BlockOutput, error) {
	return &BlockOutput{
		Block:     s.nodeview.BlockStore().LoadBlock(int64(in.Height)),
		BlockMeta: s.nodeview.BlockStore().LoadBlockMeta(int64(in.Height)),
	}, nil
}

func (s *blockchainServer) GetConsensusState(ctx context.Context, in *Empty) (*DumpConsensusStateOutput, error) {
	peerRoundState, err := s.nodeview.PeerRoundStates()
	if err != nil {
		return nil, err
	}
	return &DumpConsensusStateOutput{
		RoundState:      s.nodeview.RoundState().RoundStateSimple(),
		PeerRoundStates: peerRoundState,
	}, nil
}

func (s *blockchainServer) GetGenesis(ctx context.Context, in *Empty) (*GenesisOutput, error) {
	return &GenesisOutput{
		Genesis: s.blockchain.Genesis(),
	}, nil
}

func (s *blockchainServer) GetChainID(ctx context.Context, in *Empty) (*ChainIdOutput, error) {
	return &ChainIdOutput{
		ChainName:   s.blockchain.Genesis().ChainName(),
		ChainId:     s.blockchain.ChainID(),
		GenesisHash: s.blockchain.GenesisHash(),
	}, nil
}

//Transcation Service
func (s *TranscatorServer) GetUnconfirmedTxs(ctx context.Context, un *UnconfirmedTxsInput) (*UnconfirmedTxsOutput, error) {
	// Get all transactions for now
	transactions, err := s.nodeview.MempoolTransactions(int(un.MaxTxs))
	if err != nil {
		return nil, err
	}
	wrappedTxs := make([]*txs.Envelope, len(transactions))
	for i, tx := range transactions {
		wrappedTxs[i] = tx
	}

	return &UnconfirmedTxsOutput{
		Count: int32(len(transactions)),
		Txs:   wrappedTxs,
	}, nil
}

func (s *blockchainServer) GetBlockTxs(ctx context.Context, in *BlockInput) (*BlockTxsOutput, error) {
	
	result, err := s.GetBlock(ctx, in)
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
	return &BlockTxsOutput{
		Count: int32(len(txsBuff)),
		Txs:   txList,
	}, nil
}

func (s *TranscatorServer) BroadcastTx(ctx context.Context, txenv *Transactparam) (*Receipt, error) {

	// s.logger.Trace.Log("method", "BroadcastTx",
	// 	"tx_hash", txEnv.Hash(),
	// 	"tx", txEnv.String())
	// txhash, err := s.transact.BroadcastTx(txEnv)
	// if err != nil {
	// 	return nil, err
	// }
	return nil, nil
}

func (s *networkServer) GetNetworkInfo(context.Context, *Empty) (*NetInfoOutput, error) {
	listening := s.nodeview.IsListening()
	var contexts context.Context
	var listeners []string
	for _, listener := range s.nodeview.Listeners() {
		listeners = append(listeners, listener.String())
	}
	peers, err := s.GetPeers(contexts, nil)
	if err != nil {
		return nil, err
	}
	return &NetInfoOutput{
		Listening: listening,
		Listeners: listeners,
		Peer:      peers.Peer,
	}, nil
}

//Network service
func (s *networkServer) GetPeers(context.Context, *Empty) (*PeersOutput, error) {
	peers := make([]*rpc.Peer, s.nodeview.Peers().Size())
	for i, peer := range s.nodeview.Peers().List() {
		peers[i] = &rpc.Peer{
			NodeInfo:   peer.NodeInfo(),
			IsOutbound: peer.IsOutbound(),
		}
	}

	return &PeersOutput{
		Peer: peers,
	}, nil
}
