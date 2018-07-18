package crypto

import (
	"encoding/hex"

	"github.com/mr-tron/base58/base58"
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

func PublicKeyFromString(s string) (PublicKey, error) {
	var pb PublicKey
	bs, err := base58.Decode(s)
	if err != nil {
		return pb, err
	}

	return PublicKeyFromRawBytes(bs)
}

// PublicKeyFromRawBytes reads the raw bytes and returns an ed25519 public key.
func PublicKeyFromRawBytes(bs []byte) (PublicKey, error) {
	pb := PublicKey{
		data: publicKeyData{
			PublicKey: bs,
		},
	}

	if err := pb.check(); err != nil {
		return PublicKey{}, err
	}

	return pb, nil
}

func (pb *PublicKey) check() error {
	return nil
}

/// -------
/// CASTING

func (pb PublicKey) RawBytes() []byte {
	return pb.data.PublicKey[:]
}

// TMPubKey returns the tendermint PubKey.
func (pb PublicKey) TMPubKey() tmCrypto.PubKey {
	pk := tmCrypto.PubKeyEd25519{}
	copy(pk[:], pb.RawBytes())
	return pk
}

func (pb PublicKey) String() string {
	return hex.EncodeToString(pb.RawBytes())
}

/// ----------
/// MARSHALING

func (pb PublicKey) MarshalText() ([]byte, error) {
	str := pb.String()
	return []byte(str), nil
}

func (pb *PublicKey) UnmarshalText(bs []byte) error {
	str := string(bs)

	bs, err := hex.DecodeString(str)
	if err != nil {
		return err
	}

	p, err := PublicKeyFromRawBytes(bs)
	if err != nil {
		return err
	}

	*pb = p
	return nil
}

/// ----------
/// ATTRIBUTES

func (pb *PublicKey) IsValid() bool {
	return pb.check() == nil
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
