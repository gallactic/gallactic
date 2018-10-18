package crypto

import (
	"unsafe"

	"github.com/gallactic/gallactic/errors"
	"github.com/mr-tron/base58/base58"
	tmABCI "github.com/tendermint/tendermint/abci/types"
	tmCrypto "github.com/tendermint/tendermint/crypto"
	tmCryptoED25519 "github.com/tendermint/tendermint/crypto/ed25519"
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

// PrivateKeyFromString constructs a private key from base58 encoding text and check the prefix and checksum
func PublicKeyFromString(text string) (PublicKey, error) {
	data, err := base58.Decode(text)
	if err != nil {
		return PublicKey{}, e.Errorf(e.ErrInvalidPublicKey, "%v", err.Error())
	}

	err = validateChecksum(data)
	if err != nil {
		return PublicKey{}, e.Errorf(e.ErrInvalidPublicKey, err.Error())
	}

	err = validatePrefix(data, prefixPublicKey)
	if err != nil {
		return PublicKey{}, e.Errorf(e.ErrInvalidPublicKey, err.Error())
	}

	bs := data[2 : ed25519.PublicKeySize+2]
	return PublicKeyFromRawBytes(bs)
}

// PublicKeyFromRawBytes reads the raw bytes and returns an ed25519 public key.
func PublicKeyFromRawBytes(bs []byte) (PublicKey, error) {
	/// Check for empty value
	if len(bs) == 0 {
		return PublicKey{}, nil
	}

	pb := PublicKey{
		data: publicKeyData{
			PublicKey: bs,
		},
	}

	if err := pb.EnsureValid(); err != nil {
		return PublicKey{}, err
	}

	return pb, nil
}

/// -------
/// CASTING

// RawBytes returns the ed25519 raw bytes of the public key
func (pb PublicKey) RawBytes() []byte {
	return pb.data.PublicKey[:]
}

// String return the base58 encoding text of the public key with the prefix and checksum
func (pb PublicKey) String() string {
	if len(pb.data.PublicKey) == 0 {
		return ""
	}

	prefix := prefixPublicKey
	data := make([]byte, 0, ed25519.PublicKeySize+6)
	data = append(data, (*[2]byte)(unsafe.Pointer(&prefix))[:]...)
	data = append(data, pb.data.PublicKey...)
	chksum := checksum(data)
	data = append(data, chksum[:]...)

	return base58.Encode(data)
}

func (pb PublicKey) ABCIPubKey() tmABCI.PubKey {
	return tmABCI.PubKey{
		Type: tmABCI.PubKeyEd25519,
		Data: pb.RawBytes(),
	}
}

// TMPubKey returns the tendermint PubKey.
func (pb PublicKey) TMPubKey() tmCrypto.PubKey {
	pk := tmCryptoED25519.PubKeyEd25519{}
	copy(pk[:], pb.RawBytes())
	return pk
}

/// ----------
/// MARSHALING

func (pb PublicKey) MarshalAmino() ([]byte, error) {
	return pb.RawBytes(), nil
}

func (pb *PublicKey) UnmarshalAmino(bs []byte) error {
	p, err := PublicKeyFromRawBytes(bs)
	if err != nil {
		return err
	}

	*pb = p
	return nil
}

func (pb PublicKey) MarshalText() ([]byte, error) {
	return []byte(pb.String()), nil
}

func (pb *PublicKey) UnmarshalText(text []byte) error {
	/// Unmarshal empty value
	if len(text) == 0 {
		return nil
	}

	p, err := PublicKeyFromString(string(text))
	if err != nil {
		return err
	}

	*pb = p
	return nil
}

/// ----------
/// ATTRIBUTES

func (pb *PublicKey) EnsureValid() error {
	bs := pb.RawBytes()
	if len(bs) != ed25519.PublicKeySize {
		return e.Errorf(e.ErrInvalidPublicKey, "Public key should be %v bytes but it is %v bytes", ed25519.PublicKeySize, len(bs))
	}
	return nil
}

func (pb PublicKey) Verify(msg []byte, signature Signature) bool {
	return ed25519.Verify(pb.RawBytes(), Sha3(msg), signature.RawBytes())
}

func (pb PublicKey) AccountAddress() Address {
	tmPubKey := new(tmCryptoED25519.PubKeyEd25519)
	copy(tmPubKey[:], pb.RawBytes())
	hash := tmPubKey.Address()
	addr, _ := addressFromHash(hash, prefixAccountAddress)

	return addr
}

func (pb PublicKey) ValidatorAddress() Address {
	tmPubKey := new(tmCryptoED25519.PubKeyEd25519)
	copy(tmPubKey[:], pb.RawBytes())
	hash := tmPubKey.Address()
	addr, _ := addressFromHash(hash, prefixValidatorAddress)

	return addr
}
