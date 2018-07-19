package crypto

import (
	"encoding/hex"

	"github.com/gallactic/gallactic/errors"
	tmCrypto "github.com/tendermint/tendermint/crypto"
	"golang.org/x/crypto/ed25519"
)

// PublicKey
type PublicKey struct {
	data publicKeyData
}

type publicKeyData struct {
	PublicKey []byte
}

/// ------------
/// CONSTRUCTORS

func PublicKeyFromString(text string) (PublicKey, error) {
	var pb PublicKey
	if err := pb.UnmarshalText([]byte(text)); err != nil {
		return PublicKey{}, err
	}

	return pb, nil
}

// PublicKeyFromRawBytes reads the raw bytes and returns an ed25519 public key.
func PublicKeyFromRawBytes(bs []byte) (PublicKey, error) {
	var pb PublicKey
	if err := pb.UnmarshalAmino(bs); err != nil {
		return PublicKey{}, err
	}

	return pb, nil
}

/// -------
/// CASTING

func (pb PublicKey) RawBytes() []byte {
	return pb.data.PublicKey[:]
}

func (pb PublicKey) String() string {
	return hex.EncodeToString(pb.RawBytes())
}

// TMPubKey returns the tendermint PubKey.
func (pb PublicKey) TMPubKey() tmCrypto.PubKey {
	pk := tmCrypto.PubKeyEd25519{}
	copy(pk[:], pb.RawBytes())
	return pk
}

/// ----------
/// MARSHALING

func (pb PublicKey) MarshalAmino() ([]byte, error) {
	return pb.data.PublicKey, nil
}

func (pb *PublicKey) UnmarshalAmino(bs []byte) error {
	pb.data.PublicKey = bs
	if err := pb.EnsureValid(); err != nil {
		return err
	}

	return nil
}

func (pb PublicKey) MarshalText() ([]byte, error) {
	return []byte(pb.String()), nil
}

func (pb *PublicKey) UnmarshalText(text []byte) error {
	bs, err := hex.DecodeString(string(text))
	if err != nil {
		return err
	}

	return pb.UnmarshalAmino(bs)
}

/// ----------
/// ATTRIBUTES

func (pb *PublicKey) EnsureValid() error {
	bs := pb.RawBytes()
	if len(bs) != ed25519.PublicKeySize {
		return e.Errorf(e.ErrInvalidPublicKey, "PublicKey should be %v bytes but it is %v bytes", ed25519.PublicKeySize, len(bs))
	}
	return nil
}

func (pb PublicKey) Verify(msg []byte, signature Signature) bool {
	return ed25519.Verify(pb.data.PublicKey, msg, signature.RawBytes())
}

func (pb PublicKey) AccountAddress() Address {
	tmPubKey := new(tmCrypto.PubKeyEd25519)
	copy(tmPubKey[:], pb.RawBytes())
	hash := tmPubKey.Address()
	address, _ := addressFromHash(hash, accountAddress)

	return address
}

func (pb PublicKey) ValidatorAddress() Address {
	tmPubKey := new(tmCrypto.PubKeyEd25519)
	copy(tmPubKey[:], pb.RawBytes())
	hash := tmPubKey.Address()
	address, _ := addressFromHash(hash, validatorAddress)

	return address
}
