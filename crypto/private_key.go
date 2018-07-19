package crypto

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"io"

	"github.com/gallactic/gallactic/errors"
	"golang.org/x/crypto/ed25519"
)

type PrivateKey struct {
	data privateKeyData
}

type privateKeyData struct {
	PrivateKey []byte
}

/// ------------
/// CONSTRUCTORS

func PrivateKeyFromString(text string) (PrivateKey, error) {
	var pv PrivateKey
	if err := pv.UnmarshalText([]byte(text)); err != nil {
		return PrivateKey{}, err
	}

	return pv, nil
}

func PrivateKeyFromRawBytes(bs []byte) (PrivateKey, error) {
	var pv PrivateKey
	if err := pv.UnmarshalAmino(bs); err != nil {
		return PrivateKey{}, err
	}

	return pv, nil
}

func PrivateKeyFromSecret(secret string) PrivateKey {
	hasher := sha256.New()
	hasher.Write(([]byte)(secret))

	return GeneratePrivateKey(bytes.NewBuffer(hasher.Sum(nil)))
}

func GeneratePrivateKey(random io.Reader) PrivateKey {
	if random == nil {
		random = rand.Reader
	}
	// No error from a buffer
	_, privKey, _ := ed25519.GenerateKey(random)
	pv, _ := PrivateKeyFromRawBytes(privKey)
	return pv
}

/// -------
/// CASTING

func (pv PrivateKey) RawBytes() []byte {
	return pv.data.PrivateKey
}

func (pv PrivateKey) String() string {
	return hex.EncodeToString(pv.RawBytes())
}

/// ----------
/// MARSHALING

func (pv PrivateKey) MarshalAmino() ([]byte, error) {
	return pv.data.PrivateKey, nil
}

func (pv *PrivateKey) UnmarshalAmino(bs []byte) error {
	pv.data.PrivateKey = bs
	if err := pv.EnsureValid(); err != nil {
		return err
	}

	return nil
}

func (pv PrivateKey) MarshalText() ([]byte, error) {
	return []byte(pv.String()), nil
}

func (pv *PrivateKey) UnmarshalText(text []byte) error {
	bs, err := hex.DecodeString(string(text))
	if err != nil {
		return err
	}

	return pv.UnmarshalAmino(bs)
}

/// -------
/// METHODS

func (pv *PrivateKey) EnsureValid() error {
	bs := pv.RawBytes()
	if len(bs) != ed25519.PrivateKeySize {
		return e.Errorf(e.ErrInvalidPrivateKey, "PrivateKey should be %v bytes but it is %v bytes", ed25519.PrivateKeySize, len(bs))
	}
	_, derivedPrivateKey, err := ed25519.GenerateKey(bytes.NewBuffer(bs))
	if err != nil {
		return e.Error(e.ErrInvalidPrivateKey)
	}
	if !bytes.Equal(derivedPrivateKey, bs) {
		return e.Error(e.ErrInvalidPrivateKey)
	}
	return nil
}

func (pv PrivateKey) Sign(msg []byte) (Signature, error) {
	privKey := ed25519.PrivateKey(pv.data.PrivateKey)
	return SignatureFromRawBytes(ed25519.Sign(privKey, msg))

}

func (pv PrivateKey) PublicKey() PublicKey {
	publicKey, _ := PublicKeyFromRawBytes(pv.RawBytes()[32:])
	return publicKey
}
