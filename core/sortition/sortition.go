package sortition

import (
	"github.com/gallactic/gallactic/core/state"
	"github.com/gallactic/gallactic/crypto"
	"github.com/hyperledger/burrow/logging"
)

type Sortition interface {
	Address() crypto.Address
	Evaluate(blockHeight uint64, prevBlockHash []byte)
	Verify(prevBlockHash []byte, publicKey crypto.PublicKey, info uint64, proof []byte) bool
}

type ITransactor interface {
	//BroadcastTx(tx Tx.Payload) (*txs.Receipt, error)
}

type sortition struct {
	signer        crypto.Signer
	vrf           VRF
	transactor    ITransactor
	chainID       string
	validatorPool state.ValidatorPool
	logger        *logging.Logger
}

func NewSortition(signer crypto.Signer, chainID string, logger *logging.Logger) Sortition {
	return &sortition{
		signer:     signer,
		vrf:        NewVRF(signer),
		chainID:    chainID,
		logger:     logger,
		transactor: nil,
	}
}

func (s *sortition) SetTransactor(transactor ITransactor) {
	s.transactor = transactor
}

func (s *sortition) SetValidaorPool(validatorPool ValidatorPool) {
	s.validatorPool = validatorPool
}

// Evaluate return the vrf for self chossing to be a validator
func (s *sortition) Evaluate(blockHeight uint64, prevBlockHash []byte) {
	totalStake, valStake := s.getTotalStake(s.Address())

	s.vrf.SetMax(totalStake)

	index, _ := s.vrf.Evaluate(prevBlockHash)

	if index < valStake {

		s.logger.InfoMsg("This validator is choosen to be in set at height %v", blockHeight)
		/*
				tx := txs.NewSortitionTx(
					s.PublicKey(),
					blockHeight,
					index,
					proof)

				tx.Signature, _ = acm.ChainSign(s, s.chainID, tx)

			if s.transactor != nil {
				s.transactor.BroadcastTx(tx)
			}
		*/
	}
}

func (s *sortition) Verify(prevBlockHash []byte, publicKey crypto.PublicKey, index uint64, proof []byte) bool {

	totalStake, valStake := s.getTotalStake(publicKey.Address())

	// Note: totalStake can be changed by time on verifying
	// So we calculate the index again
	s.vrf.SetMax(totalStake)

	index2, result := s.vrf.Verify(prevBlockHash, publicKey, proof)

	if result == false {
		return false
	}

	return index2 < valStake
}

func (s *sortition) Address() crypto.Address {
	return s.Address()
}

func (s *sortition) getTotalStake(addr crypto.Address) (totalStake uint64, validatorStake uint64) {
	totalStake = 0
	validatorStake = 0

	s.validatorPool.IterateValidator(func(validator *Validator) (stop bool) {
		totalStake += validator.Stake()

		if address == validator.Address() {
			validatorStake = validator.Stake()
		}

		return false
	})

	return
}
