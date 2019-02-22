package grpc

import (
	"context"
	"encoding/hex"
	"encoding/json"

	"github.com/gallactic/gallactic/common/binary"
	"github.com/gallactic/gallactic/core/account"
	"github.com/gallactic/gallactic/core/blockchain"
	"github.com/gallactic/gallactic/core/consensus/tendermint/p2p"
	"github.com/gallactic/gallactic/core/consensus/tendermint/query"
	"github.com/gallactic/gallactic/core/state"
	"github.com/gallactic/gallactic/core/validator"
	"github.com/gallactic/gallactic/crypto"
	"github.com/gallactic/gallactic/txs"
	"github.com/gallactic/gallactic/version"
	pb "github.com/gallactic/gallactic/www/grpc/proto3"
	log "github.com/inconshreveable/log15"
	consensusTypes "github.com/tendermint/tendermint/consensus/types"
	net "github.com/tendermint/tendermint/p2p"
	tmRPC "github.com/tendermint/tendermint/rpc/core"
	tmTypes "github.com/tendermint/tendermint/types"
)

// MaxBlockLookback constant
const MaxBlockLookback = 1000

type blockchainService struct {
	nodeview   *query.NodeView
	blockchain *blockchain.Blockchain
	state      *state.State
}

var _ pb.BlockChainServer = &blockchainService{}

func (s *blockchainService) State() *state.State {
	return s.state
}

func NewBlockchainService(blockchain *blockchain.Blockchain, nview *query.NodeView) *blockchainService {
	return &blockchainService{
		blockchain: blockchain,
		nodeview:   nview,
		state:      blockchain.State(),
	}
}

// Blockchain Service
func (as *blockchainService) GetAccount(ctx context.Context, param *pb.AddressRequest) (*pb.AccountResponse, error) {
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

func (as *blockchainService) GetAccounts(ctx context.Context, in *pb.Empty) (*pb.AccountsResponse, error) {
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

func (vs *blockchainService) GetValidator(ctx context.Context, param *pb.AddressRequest) (*pb.ValidatorResponse, error) {
	addr, err := crypto.AddressFromString(param.Address)
	if err != nil {
		return nil, err
	}
	val, err := vs.state.GetValidator(addr)
	if err != nil {
		return nil, err
	}
	valInfo := vs.toValidatorInfo(val)
	return &pb.ValidatorResponse{Validator: valInfo}, nil
}

func (vs *blockchainService) GetValidators(context.Context, *pb.Empty) (*pb.ValidatorsResponse, error) {
	validators := make([]*pb.ValidatorInfo, 0)

	vs.state.IterateValidators(func(val *validator.Validator) (stop bool) {
		if val != nil {
			valInfo := vs.toValidatorInfo(val)
			validators = append(validators, valInfo)
		}
		return false
	})
	return &pb.ValidatorsResponse{
		Validators:  validators,
		BlockHeight: vs.blockchain.LastBlockHeight(),
	}, nil
}

func (s *blockchainService) GetStorage(ctx context.Context, storage *pb.StorageRequest) (*pb.StorageResponse, error) {
	var storageItems []pb.StorageItem

	storageaddr, err := crypto.AddressFromString(storage.Address)
	if err != nil {
		return nil, err
	}

	s.state.IterateStorage(storageaddr, func(key, value binary.Word256) (stop bool) {
		storageItems = append(storageItems, pb.StorageItem{Key: key.UnpadLeft(), Value: value.UnpadLeft()})
		return false
	})
	return &pb.StorageResponse{
		StorageItems: storageItems,
	}, nil

}

func (s *blockchainService) GetStorageAt(ctx context.Context, storage *pb.StorageAtRequest) (*pb.StorageAtResponse, error) {
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

func (s *blockchainService) GetStatus(ctx context.Context, in *pb.Empty) (*pb.StatusResponse, error) {
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
	ni := new(p2p.GNodeInfo)
	tmni := s.nodeview.NodeInfo().(net.DefaultNodeInfo)
	ni.ID_ = tmni.ID_
	ni.Network = tmni.Network
	ni.ProtocolVersion = tmni.ProtocolVersion
	ni.Version = tmni.Version
	ni.Channels = tmni.Channels
	ni.ListenAddr = tmni.ListenAddr
	ni.Moniker = tmni.Moniker
	return &pb.StatusResponse{
		NodeInfo:          *ni,
		GenesisHash:       s.blockchain.GenesisHash(),
		PubKey:            publicKey,
		LatestBlockHash:   latestBlockHash,
		LatestBlockHeight: latestHeight,
		LatestBlockTime:   latestBlockTime,
		NodeVersion:       version.Version,
	}, err
}

func (s *blockchainService) GetBlock(ctx context.Context, req *pb.BlockRequest) (*pb.BlockResponse, error) {
	height := int64(req.Height)
	if height == 0 {
		height = s.nodeview.BlockStore().Height()

	}

	return &pb.BlockResponse{
		Block: s.toBlockInfo(height),
	}, nil

}

func (s *blockchainService) GetBlocks(ctx context.Context, blocks *pb.BlocksRequest) (*pb.BlocksResponse, error) {
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
	var pbBlocks []pb.BlockInfo
	for height := blocks.MaxHeight; height >= blocks.MinHeight; height-- {
		bl := s.toBlockInfo(int64(height))
		pbBlocks = append(pbBlocks, *bl)
	}

	return &pb.BlocksResponse{
		Blocks: pbBlocks,
	}, nil

}

func (s *blockchainService) GetGenesis(context.Context, *pb.Empty) (*pb.GenesisResponse, error) {
	gen := s.blockchain.Genesis()
	return &pb.GenesisResponse{
		Genesis: gen,
	}, nil
}

func (s *blockchainService) GetChainID(context.Context, *pb.Empty) (*pb.ChainResponse, error) {
	return &pb.ChainResponse{
		ChainName:   s.blockchain.Genesis().ChainName(),
		ChainId:     s.blockchain.ChainID(),
		GenesisHash: s.blockchain.GenesisHash(),
	}, nil

}

func (s *blockchainService) GetLatestBlock(context.Context, *pb.Empty) (*pb.BlockResponse, error) {
	latestHeight := s.blockchain.LastBlockHeight()
	bl := s.toBlockInfo(int64(latestHeight))
	return &pb.BlockResponse{
		Block: bl,
	}, nil
}

func (s *blockchainService) GetBlockchainInfo(ctx context.Context, blockinfo *pb.Empty) (*pb.BlockchainInfoResponse, error) {
	res := &pb.BlockchainInfoResponse{
		LastBlockHeight: s.blockchain.LastBlockHeight(),
		LastBlockHash:   s.blockchain.LastBlockHash(),
		LastBlockTime:   s.blockchain.LastBlockTime(),
	}
	return res, nil
}

func (s *blockchainService) GetConsensusState(context.Context, *pb.Empty) (*pb.ConsensusResponse, error) {
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

func (s *blockchainService) GetBlockTxs(ctx context.Context, block *pb.BlockRequest) (*pb.BlockTxsResponse, error) {
	result, err := s.GetBlock(ctx, block)
	if err != nil {
		return nil, err
	}
	txList := result.Block.Txs

	return &pb.BlockTxsResponse{
		Count: int32(len(txList)),
		Txs:   txList,
	}, nil

}

func (s *blockchainService) GetTx(ctx context.Context, req *pb.TxRequest) (*pb.TxResponse, error) {
	hash, err := hex.DecodeString(req.Hash)
	if err != nil {
		return nil, err
	}

	tx, err := tmRPC.Tx(hash, false)
	if err != nil {
		return nil, err
	}

	return &pb.TxResponse{
		Tx: s.toTxInfo(tx.Tx),
	}, nil

}

//
func (vs *blockchainService) toValidatorInfo(val *validator.Validator) *pb.ValidatorInfo {
	return &pb.ValidatorInfo{
		Address:  val.Address().String(),
		PubKey:   val.PublicKey().String(),
		Power:    val.Power(),
		Stake:    val.Stake(),
		Sequence: val.Sequence(),
	}
}

func (s *blockchainService) toBlockInfo(blockheight int64) *pb.BlockInfo {
	pbBlock := new(pb.BlockInfo)
	blockmeta := s.nodeview.BlockStore().LoadBlockMeta(blockheight)
	block := s.nodeview.BlockStore().LoadBlock(blockheight)
	if blockmeta == nil || block == nil {
		log.Error("Invalid block height", "height", blockheight)
		return pbBlock
	}
	pbBlock.Header.BlockHash = blockmeta.BlockID.Hash.Bytes()
	pbBlock.Header.Time = blockmeta.Header.Time
	pbBlock.Header.TotalTxs = blockmeta.Header.TotalTxs
	pbBlock.Header.Version.App = blockmeta.Header.Version.App.Uint64()
	pbBlock.Header.Version.Block = blockmeta.Header.Version.Block.Uint64()
	pbBlock.Header.ChainID = blockmeta.Header.ChainID
	pbBlock.Header.Height = blockmeta.Header.Height
	pbBlock.Header.NumTxs = blockmeta.Header.NumTxs
	pbBlock.Header.LastBlockId = blockmeta.Header.LastBlockID.Hash // ignoring PartSetHeader
	pbBlock.Header.LastCommitHash = blockmeta.Header.LastCommitHash.Bytes()
	pbBlock.Header.DataHash = blockmeta.Header.DataHash.Bytes()
	pbBlock.Header.ValidatorsHash = blockmeta.Header.ValidatorsHash.Bytes()
	pbBlock.Header.NextValidatorsHash = blockmeta.Header.NextValidatorsHash.Bytes()
	pbBlock.Header.ConsensusHash = blockmeta.Header.ConsensusHash.Bytes()
	pbBlock.Header.AppHash = blockmeta.Header.AppHash.Bytes()
	pbBlock.Header.LastResultsHash = blockmeta.Header.LastResultsHash.Bytes()
	pbBlock.Header.EvidenceHash = blockmeta.Header.EvidenceHash.Bytes()
	valAddr, _ := crypto.ValidatorAddress(blockmeta.Header.ProposerAddress)
	pbBlock.Header.ProposerAddress = valAddr.String()

	for _, tx := range block.Data.Txs {
		txInfo := s.toTxInfo(tx)
		pbBlock.Txs = append(pbBlock.Txs, *txInfo)
	}
	pbBlock.LastCommitInfo.BlockHash = block.LastCommit.BlockID.Hash.Bytes()

	for _, v := range block.LastCommit.Precommits {
		if v == nil {
			continue
		}

		var vote pb.VoteInfo
		valAddr, _ := crypto.ValidatorAddress(v.ValidatorAddress)
		vote.Round = int32(v.Round)
		vote.Time = v.Timestamp
		vote.ValidatorAddress = valAddr.String()
		vote.Height = v.Height
		vote.Signature = v.Signature

		pbBlock.LastCommitInfo.Votes = append(pbBlock.LastCommitInfo.Votes, &vote)
	}

	for _, ev := range block.Evidence.Evidence {
		var evidence pb.EvidenceInfo
		valAddr, _ := crypto.ValidatorAddress(ev.Address())
		evidence.Height = ev.Height()
		evidence.Address = valAddr.String()
		pbBlock.ByzantineValidators = append(pbBlock.ByzantineValidators, evidence)
	}
	return pbBlock
}

func (s *blockchainService) toTxInfo(tx tmTypes.Tx) *pb.TxInfo {
	txInfo := new(pb.TxInfo)
	var env txs.Envelope
	if err := env.Decode(tx); err != nil {
		log.Error("Unable to decode transaction",
			"error", err)
	}

	js, _ := json.Marshal(env)
	txInfo.Hash = hex.EncodeToString(tx.Hash())
	txInfo.Envelope = string(js)
	return txInfo
}
