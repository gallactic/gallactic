package sortition

import (
	"encoding/hex"
	"math/big"

	"github.com/gallactic/gallactic/core/evm/sha3"
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
	// sign the hashed block height
	sig, err := vrf.signer.Sign(m)

	if err != nil {
		return 0, nil
	}

	proof = make([]byte, 0)
	addrBytes := vrf.signer.Address().RawBytes()
	sigBytes := sig.RawBytes()
	proof = append(proof, addrBytes...)
	proof = append(proof, sigBytes...)

	index = vrf.getIndex(sigBytes)

	return index, proof
}

// Verify ensure the proof is valid
func (vrf *VRF) Verify(m []byte, publicKey crypto.PublicKey, proof []byte) (index uint64, result bool) {
	address, err := crypto.AddressFromRawBytes(proof[0:crypto.AddressSize])
	if err != nil {
		return 0, false
	}

	sig, err := crypto.SignatureFromRawBytes(proof[crypto.AddressSize:])
	if err != nil {
		return 0, false
	}

	// Verify address
	if !address.Verify(publicKey) {
		return 0, false
	}

	// Verify signature (proof)
	if !publicKey.Verify(m, sig) {
		return 0, false
	}

	index = vrf.getIndex(sig.RawBytes())

	return index, true
}

func (vrf *VRF) getIndex(sig []byte) uint64 {
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
}
