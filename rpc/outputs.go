package rpc

import (
	"encoding/json"

	"time"

	"github.com/gallactic/gallactic/binary"
	"github.com/gallactic/gallactic/core/account"
	"github.com/gallactic/gallactic/core/genesis"
	"github.com/gallactic/gallactic/core/validator"
	"github.com/gallactic/gallactic/crypto"
	"github.com/gallactic/gallactic/txs"
	"github.com/tendermint/go-amino"
	consensusTypes "github.com/tendermint/tendermint/consensus/types"
	"github.com/tendermint/tendermint/p2p"
	"github.com/tendermint/tendermint/rpc/core/types"
	tmTypes "github.com/tendermint/tendermint/types"
)

// When using Tendermint types like Block and Vote we are forced to wrap the outer object and use amino marshalling
var aminoCodec = amino.NewCodec()

func init() {
	core_types.RegisterAmino(aminoCodec)
}

type ResultGetStorage struct {
	Key   binary.HexBytes
	Value binary.HexBytes
}

type ResultListAccounts struct {
	BlockHeight uint64
	Accounts    []*account.Account
}

type ResultDumpStorage struct {
	StorageRoot  binary.HexBytes
	StorageItems []StorageItem
}

type StorageItem struct {
	Key   binary.HexBytes
	Value binary.HexBytes
}

type ResultListBlocks struct {
	LastHeight uint64
	BlockMetas []*tmTypes.BlockMeta
}

type ResultGetBlock struct {
	BlockMeta *BlockMeta
	Block     *Block
}

type BlockMeta struct {
	*tmTypes.BlockMeta
}

func (bm BlockMeta) MarshalJSON() ([]byte, error) {
	return aminoCodec.MarshalJSON(bm.BlockMeta)
}

func (bm *BlockMeta) UnmarshalJSON(data []byte) (err error) {
	return aminoCodec.UnmarshalJSON(data, &bm.BlockMeta)
}

// Needed for go-amino handling of interface types
type Block struct {
	*tmTypes.Block
}

func (b Block) MarshalJSON() ([]byte, error) {
	return aminoCodec.MarshalJSON(b.Block)
}

func (b *Block) UnmarshalJSON(data []byte) (err error) {
	return aminoCodec.UnmarshalJSON(data, &b.Block)
}

type ResultStatus struct {
	NodeInfo          p2p.NodeInfo
	GenesisHash       binary.HexBytes
	PubKey            crypto.PublicKey
	LatestBlockHash   binary.HexBytes
	LatestBlockHeight uint64
	LatestBlockTime   int64
	NodeVersion       string
}

type ResultLastBlockInfo struct {
	LastBlockHeight uint64
	LastBlockTime   time.Time
	LastBlockHash   binary.HexBytes
}

type ResultChainId struct {
	ChainName   string
	ChainId     string
	GenesisHash binary.HexBytes
}

type Peer struct {
	NodeInfo   p2p.NodeInfo
	IsOutbound bool
}

type ResultNetInfo struct {
	Listening bool
	Listeners []string
	Peers     []*Peer
}

type ResultListValidators struct {
	BlockHeight         uint64
	BondedValidators    []*validator.Validator
	UnbondingValidators []*validator.Validator
}

type ResultDumpConsensusState struct {
	RoundState      consensusTypes.RoundStateSimple
	PeerRoundStates []*consensusTypes.PeerRoundState
}

type ResultPeers struct {
	Peers []*Peer
}

type ResultGetAccount struct {
	Account *account.Account
}

type AccountHumanReadable struct {
	Address     crypto.Address
	PublicKey   crypto.PublicKey
	Sequence    uint64
	Balance     uint64
	Code        []string
	StorageRoot string
	Permissions []string
	Roles       []string
}

type ResultGetAccountHumanReadable struct {
	Account *AccountHumanReadable
}

type ResultBroadcastTx struct {
	txs.Receipt
}

func (rbt ResultBroadcastTx) MarshalJSON() ([]byte, error) {
	return json.Marshal(rbt.Receipt)
}

func (rbt ResultBroadcastTx) UnmarshalJSON(data []byte) (err error) {
	return json.Unmarshal(data, &rbt.Receipt)
}

type ResultListUnconfirmedTxs struct {
	NumTxs int
	Txs    []*txs.Envelope
}

type ResultGenesis struct {
	Genesis *genesis.Genesis
}
