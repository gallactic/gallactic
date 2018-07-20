// Copyright 2017 Monax Industries Limited
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package rpc

import (
	"context"

	"encoding/json"
	"fmt"
	"github.com/gallactic/gallactic/common/binary"
	"github.com/gallactic/gallactic/core/account"
	"github.com/gallactic/gallactic/core/blockchain"
	"github.com/gallactic/gallactic/core/consensus/tendermint/query"
	"github.com/gallactic/gallactic/core/execution"
	"github.com/gallactic/gallactic/core/state"
	"github.com/gallactic/gallactic/core/validator"
	"github.com/gallactic/gallactic/crypto"
	"github.com/gallactic/gallactic/txs"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/logging/structure"
	tmTypes "github.com/tendermint/tendermint/types"
	"time"
)

// Magic! Should probably be configurable, but not shouldn't be so huge we
// end up DoSing ourselves.
const MaxBlockLookback = 1000
const AccountsRingMutexCount = 100

// Base service that provides implementation for all underlying RPC methods
type Service struct {
	ctx        context.Context
	state      *state.State
	blockchain *blockchain.Blockchain
	transactor *execution.Transactor
	logger     *logging.Logger
	nodeView   *query.NodeView
}

func NewService(ctx context.Context, blockchain *blockchain.Blockchain,
	transactor *execution.Transactor, nView *query.NodeView, logger *logging.Logger) *Service {

	return &Service{
		ctx:        ctx,
		state:      blockchain.State(),
		blockchain: blockchain,
		transactor: transactor,
		logger:     logger.With(structure.ComponentKey, "Service"),
		nodeView:   nView,
	}
}

func (s *Service) Transactor() *execution.Transactor {
	return s.transactor
}

func (s *Service) State() *state.State {
	return s.state
}

func (s *Service) BlockchainInfo() *blockchain.Blockchain {
	return s.blockchain
}

func (s *Service) ListUnconfirmedTxs(maxTxs int) (*UnconfirmedTxsOutput, error) {
	// Get all transactions for now
	transactions, err := s.nodeView.MempoolTransactions(maxTxs)
	if err != nil {
		return nil, err
	}
	wrappedTxs := make([]*txs.Envelope, len(transactions))
	for i, tx := range transactions {
		wrappedTxs[i] = tx
	}
	return &UnconfirmedTxsOutput{
		Count: len(transactions),
		Txs:   wrappedTxs,
	}, nil
}
func (s *Service) ListBlockTxs(height uint64) (*BlockTxsOutput, error) {
	result, err := s.GetBlock(height)
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
		Count: len(txsBuff),
		Txs:   txList,
	}, nil
}
func (s *Service) Status() (*StatusOutput, error) {
	latestHeight := s.blockchain.LastBlockHeight()
	var (
		latestBlockMeta *tmTypes.BlockMeta
		latestBlockHash []byte
		latestBlockTime int64
	)
	if latestHeight != 0 {
		latestBlockMeta = s.nodeView.BlockStore().LoadBlockMeta(int64(latestHeight))
		latestBlockHash = latestBlockMeta.Header.Hash()
		latestBlockTime = latestBlockMeta.Header.Time.UnixNano()
	}
	publicKey, err := s.nodeView.PrivValidatorPublicKey()
	if err != nil {
		return nil, err
	}
	return &StatusOutput{
		NodeInfo:          s.nodeView.NodeInfo(),
		GenesisHash:       s.blockchain.GenesisHash(),
		PubKey:            publicKey,
		LatestBlockHash:   latestBlockHash,
		LatestBlockHeight: latestHeight,
		LatestBlockTime:   latestBlockTime,
		// TODO Ahmad
		//NodeVersion:       project.History.CurrentVersion().String(),
	}, nil
}

func (s *Service) ChainIdentifiers() (*ChainIdOutput, error) {
	return &ChainIdOutput{
		ChainName:   s.blockchain.Genesis().ChainName(),
		ChainId:     s.blockchain.ChainID(),
		GenesisHash: s.blockchain.GenesisHash(),
	}, nil
}

func (s *Service) Peers() (*PeersOutput, error) {
	peers := make([]*Peer, s.nodeView.Peers().Size())
	for i, peer := range s.nodeView.Peers().List() {
		peers[i] = &Peer{
			NodeInfo:   peer.NodeInfo(),
			IsOutbound: peer.IsOutbound(),
		}
	}
	return &PeersOutput{
		Peers: peers,
	}, nil
}

func (s *Service) NetInfo() (*NetInfoOutput, error) {
	listening := s.nodeView.IsListening()
	var listeners []string
	for _, listener := range s.nodeView.Listeners() {
		listeners = append(listeners, listener.String())
	}
	peers, err := s.Peers()
	if err != nil {
		return nil, err
	}
	return &NetInfoOutput{
		Listening: listening,
		Listeners: listeners,
		Peers:     peers.Peers,
	}, nil
}

func (s *Service) Genesis() *GenesisOutput {
	return &GenesisOutput{
		Genesis: s.blockchain.Genesis(),
	}
}

func (s *Service) GetAccount(address crypto.Address) (*AccountOutput, error) {
	acc := s.state.GetAccount(address)
	if acc == nil {
		return nil, nil //TODO we should return a proper error!
	}
	return &AccountOutput{Account: acc}, nil
}

func (s *Service) ListAccounts(predicate func(*account.Account) bool) (*AccountsOutput, error) {
	accounts := make([]*account.Account, 0)
	s.state.IterateAccounts(func(acc *account.Account) (stop bool) {
		if predicate(acc) {
			accounts = append(accounts, acc)
		}
		return
	})

	return &AccountsOutput{
		BlockHeight: s.blockchain.LastBlockHeight(),
		Accounts:    accounts,
	}, nil
}

func (s *Service) GetStorage(address crypto.Address, key []byte) (*StorageOutput, error) {
	acc := s.state.GetAccount(address)
	if acc == nil {
		return nil, fmt.Errorf("UnknownAddress: %s", address)
	}

	value, err := s.state.GetStorage(address, binary.LeftPadWord256(key))
	if err != nil {
		return nil, err
	}
	if value == binary.Zero256 {
		return &StorageOutput{Key: key, Value: nil}, nil
	}
	return &StorageOutput{Key: key, Value: value.UnpadLeft()}, nil
}

func (s *Service) DumpStorage(address crypto.Address) (*DumpstorageOutput, error) {
	acc := s.state.GetAccount(address)
	if acc == nil {
		return nil, fmt.Errorf("UnknownAddress: %X", address)
	}

	var storageItems []StorageItem
	s.state.IterateStorage(address, func(key, value binary.Word256) (stop bool) {
		storageItems = append(storageItems, StorageItem{Key: key.UnpadLeft(), Value: value.UnpadLeft()})
		return
	})
	return &DumpstorageOutput{
		StorageRoot:  acc.StorageRoot(),
		StorageItems: storageItems,
	}, nil
}

func (s *Service) GetBlock(height uint64) (*BlockOutput, error) {
	return &BlockOutput{
		Block:     &Block{s.nodeView.BlockStore().LoadBlock(int64(height))},
		BlockMeta: &BlockMeta{s.nodeView.BlockStore().LoadBlockMeta(int64(height))},
	}, nil
}

// Returns the current blockchain height and metadata for a range of blocks
// between minHeight and maxHeight. Only returns maxBlockLookback block metadata
// from the top of the range of blocks.
// Passing 0 for maxHeight sets the upper height of the range to the current
// blockchain height.
func (s *Service) ListBlocks(minHeight, maxHeight uint64) (*BlocksOutput, error) {
	latestHeight := s.blockchain.LastBlockHeight()

	if minHeight == 0 {
		minHeight = 1
	}
	if maxHeight == 0 || latestHeight < maxHeight {
		maxHeight = latestHeight
	}
	if maxHeight > minHeight && maxHeight-minHeight > MaxBlockLookback {
		minHeight = maxHeight - MaxBlockLookback
	}

	var blockMetas []*tmTypes.BlockMeta
	for height := maxHeight; height >= minHeight; height-- {
		blockMeta := s.nodeView.BlockStore().LoadBlockMeta(int64(height))
		blockMetas = append(blockMetas, blockMeta)
	}

	return &BlocksOutput{
		LastHeight: latestHeight,
		BlockMetas: blockMetas,
	}, nil
}

func (s *Service) ListValidators() (*ValidatorsOutput, error) {
	validators := make([]*validator.Validator, 0)
	s.blockchain.State().IterateValidators(func(val *validator.Validator) (stop bool) {
		validators = append(validators, val)
		return
	})
	return &ValidatorsOutput{
		BlockHeight:         s.blockchain.LastBlockHeight(),
		BondedValidators:    validators,
		UnbondingValidators: nil,
	}, nil
}

func (s *Service) DumpConsensusState() (*DumpConsensusStateOutput, error) {
	peerRoundState, err := s.nodeView.PeerRoundStates()
	if err != nil {
		return nil, err
	}
	return &DumpConsensusStateOutput{
		RoundState:      s.nodeView.RoundState().RoundStateSimple(),
		PeerRoundStates: peerRoundState,
	}, nil
}

func (s *Service) LastBlockInfo(blockWithin string) (*LastBlockInfoOutput, error) {
	res := &LastBlockInfoOutput{
		LastBlockHeight: s.blockchain.LastBlockHeight(),
		LastBlockHash:   s.blockchain.LastBlockHash(),
		LastBlockTime:   s.blockchain.LastBlockTime(),
	}
	if blockWithin == "" {
		return res, nil
	}
	duration, err := time.ParseDuration(blockWithin)
	if err != nil {
		return nil, fmt.Errorf("could not parse blockWithin duration to determine whether to throw error: %v", err)
	}
	// Take neg abs in case caller is counting backwards (not we add later)
	if duration > 0 {
		duration = -duration
	}
	blockTimeThreshold := time.Now().Add(duration)
	if res.LastBlockTime.After(blockTimeThreshold) {
		// We've created blocks recently enough
		return res, nil
	}
	resJSON, err := json.Marshal(res)
	if err != nil {
		resJSON = []byte("<error: could not marshal last block info>")
	}
	return nil, fmt.Errorf("no block committed within the last %s (cutoff: %s), last block info: %s",
		blockWithin, blockTimeThreshold.Format(time.RFC3339), string(resJSON))
}
