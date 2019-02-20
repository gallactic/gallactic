package sortition

import (
	"github.com/gallactic/gallactic/core/state"
	"github.com/gallactic/gallactic/core/validator"
	"github.com/gallactic/gallactic/crypto"
	"github.com/gallactic/gallactic/txs"
	"github.com/gallactic/gallactic/txs/tx"
	log "github.com/inconshreveable/log15"
	tmRPC "github.com/tendermint/tendermint/rpc/core"
)

type Sortition struct {
	//transactor ITransactor
	state        *state.State
	signer       crypto.Signer
	vrf          VRF
	chainID      string
	sortitionFee uint64
	logger       log.Logger
}

func NewSortition(state *state.State, signer crypto.Signer, chainID string) *Sortition {
	return &Sortition{
		signer:  signer,
		state:   state,
		chainID: chainID,
		vrf:     NewVRF(signer),
	}
}

// Evaluate return the vrf for self choosing to be a validator
func (s *Sortition) Evaluate(blockHeight uint64, blockHash []byte) {
	addr := s.signer.Address()
	totalStake, valStake := s.getTotalStake(addr)
	s.vrf.SetMax(totalStake)
	index, proof := s.vrf.Evaluate(blockHash)

	if index < valStake {
		log.Info("This validator is chosen to be in set", "height", blockHeight, "address", addr, "stake", valStake)

		/// TODO: better way????
		val, err := s.state.GetValidator(s.Address())
		if err != nil {
			return
		}

		tx, _ := tx.NewSortitionTx(
			s.signer.Address(),
			blockHeight,
			val.Sequence()+1,
			s.sortitionFee,
			index,
			proof)

		txEnv := txs.Enclose(s.chainID, tx)
		err = txEnv.Sign(s.signer)
		if err != nil {
			return
		}

		// TODO:: better way?????
		codec := txs.NewAminoCodec()
		bs, err := codec.MarshalBinaryLengthPrefixed(txEnv)
		if err != nil {
			return
		}

		res, err := tmRPC.BroadcastTxAsync(bs)
		if err != nil {
			return
		}

		if res != nil {
			/// TODO: log result
		}
	}
}

func (s *Sortition) Verify(blockHash []byte, pb crypto.PublicKey, index uint64, proof []byte) bool {
	addr := pb.ValidatorAddress()
	totalStake, valStake := s.getTotalStake(addr)

	// Note: totalStake can be changed by time on verifying
	// So we calculate the index again
	s.vrf.SetMax(totalStake)

	index2, result := s.vrf.Verify(blockHash, pb, proof)
	if !result {
		log.Warn("Unable to verify a sortition tx", "blockhash", blockHash, "Address", addr)
		return false
	}

	return index2 < valStake
}

func (s *Sortition) Address() crypto.Address {
	return s.signer.Address()
}

func (s *Sortition) getTotalStake(addr crypto.Address) (totalStake uint64, validatorStake uint64) {
	totalStake = 0
	validatorStake = 0

	s.state.IterateValidators(func(validator *validator.Validator) (stop bool) {
		totalStake += validator.Stake()

		if addr == validator.Address() {
			validatorStake = validator.Stake()
		}

		return false
	})

	return
}
