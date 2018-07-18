package validator

import (
	"github.com/gallactic/gallactic/crypto"

	tmCrypto "github.com/tendermint/tendermint/crypto"
	tmTypes "github.com/tendermint/tendermint/types"
)

type privValidatorMemory struct {
	publicKey      tmCrypto.PubKey
	signer         tmSigner
	lastSignedInfo *LastSignedInfo
}

var _ tmTypes.PrivValidator = &privValidatorMemory{}

// Create a PrivValidator with in-memory state that takes an addressable representing the validator identity
// and a signer providing private signing for that identity.
func NewPrivValidatorMemory(pv crypto.Signer) *privValidatorMemory {
	return &privValidatorMemory{
		publicKey:      pv.PublicKey().TMPubKey(),
		signer:         asTendermintSigner(pv),
		lastSignedInfo: NewLastSignedInfo(),
	}
}

func (pvm *privValidatorMemory) GetAddress() tmTypes.Address {
	return pvm.publicKey.Address()
}

func (pvm *privValidatorMemory) GetPubKey() tmCrypto.PubKey {
	return pvm.publicKey
}

// TODO: consider persistence to disk/database to avoid double signing after a crash
func (pvm *privValidatorMemory) SignVote(chainID string, vote *tmTypes.Vote) error {
	return pvm.lastSignedInfo.SignVote(pvm.signer, chainID, vote)
}

func (pvm *privValidatorMemory) SignProposal(chainID string, proposal *tmTypes.Proposal) error {
	return pvm.lastSignedInfo.SignProposal(pvm.signer, chainID, proposal)
}

func (pvm *privValidatorMemory) SignHeartbeat(chainID string, heartbeat *tmTypes.Heartbeat) error {
	return pvm.lastSignedInfo.SignHeartbeat(pvm.signer, chainID, heartbeat)
}

func asTendermintSigner(signer crypto.Signer) tmSigner {
	return func(msg []byte) tmCrypto.Signature {
		sig, err := signer.Sign(msg)
		if err != nil {
			return nil
		}
		tmSig := tmCrypto.SignatureEd25519{}
		copy(tmSig[:], sig.RawBytes())
		return tmSig

	}
}
