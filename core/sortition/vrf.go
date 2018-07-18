package sortition

import (
	"encoding/hex"
	"math/big"

	"github.com/gallactic/gallactic/crypto"
)

type VRF struct {
	signer crypto.Signer
	max256 *big.Int
	max    *big.Int
}

func NewVRF(signer crypto.Signer) VRF {
	vrf := VRF{
		signer: signer,
		max:    big.NewInt(0),
		max256: big.NewInt(0),
	}

	decMax, _ := hex.DecodeString("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF")

	vrf.max256.SetBytes(decMax)
	vrf.max.SetBytes(decMax)

	return vrf
}

func (vrf *VRF) SetMax(max uint64) {
	vrf.max.SetUint64(max)
}

// Evaluate returns a random number between 0 and 10^18 with the proof
func (vrf *VRF) Evaluate(m []byte) (index uint64, proof []byte) {
	/*
		// sign the hashed block height
		sig, err := vrf.signer.Sign(m)

		if err != nil {
			return 0, nil
		}

		proof = make([]byte, 0)

		//////////address := vrf.signer.Address()
		address := []byte{1, 2, 3}
		proof = append(proof, address...)
		////////////proof = append(proof, sig.Bytes()...)

		///index = vrf.getIndex(sig)

		return index, proof
	*/
	return 0, nil
}

// Verify ensure the proof is valid
func (vrf *VRF) Verify(m []byte, publicKey crypto.PublicKey, proof []byte) (index uint64, result bool) {
	/*
		address, err := crypto.AddressFromBytes(proof[0:binary.Word160Length])
		if err != nil {
			return 0, false
		}

		sig, err := crypto.SignatureFromBytes(proof[binary.Word160Length+1:])
		if err != nil {
			return 0, false
		}

		// Verify address
		if publicKey.Address() != address {
			return 0, false
		}

		// Verify signature (proof)
		if !publicKey.VerifyBytes(m, sig) {
			return 0, false
		}

		index = vrf.getIndex(sig.Bytes())

		return index, true
	*/
	return 0, true
}

func (vrf *VRF) getIndex(sig []byte) uint64 {
	/*
		hash := big.NewInt(0)
		hash.SetBytes(sha3.Sha3(sig))

		// construct the numerator and denominator for normalizing the signature uint between [0, 1]
		index := big.NewInt(0)
		numerator := big.NewInt(0)

		denominator := vrf.max256

		numerator = numerator.Mul(hash, vrf.max)

		// divide numerator and denominator to get the election ratio for this block height
		index = index.Div(numerator, denominator)

		return index.Uint64()
	*/
	return 0
}
