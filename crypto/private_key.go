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
	bs, err := hex.DecodeString(text)
	if err != nil {
		return PrivateKey{}, e.Errorf(e.ErrInvalidPrivateKey, "%v", err.Error())
	}

	return PrivateKeyFromRawBytes(bs)
}

func PrivateKeyFromRawBytes(bs []byte) (PrivateKey, error) {
	/// Check for empty private key
	if len(bs) == 0 {
		return PrivateKey{}, nil
	}
	
	pv := PrivateKey{
		data: privateKeyData{
			PrivateKey: bs,
		},
	}

	if err := pv.EnsureValid(); err != nil {
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
	return pv.RawBytes(), nil
}

func (pv *PrivateKey) UnmarshalAmino(bs []byte) error {
	p, err := PrivateKeyFromRawBytes(bs)
	if err != nil {
		return err
	}

	*pv = p
	return nil
}

func (pv PrivateKey) MarshalText() ([]byte, error) {
	return []byte(pv.String()), nil
}

func (pv *PrivateKey) UnmarshalText(text []byte) error {
	p, err := PrivateKeyFromString(string(text))
	if err != nil {
		return err
	}

	*pv = p
	return nil
}

/// -------
/// METHODS

func (pv *PrivateKey) EnsureValid() error {
	bs := pv.RawBytes()
	if len(bs) != ed25519.PrivateKeySize {
		return e.Errorf(e.ErrInvalidPrivateKey, "Private key should be %v bytes but it is %v bytes", ed25519.PrivateKeySize, len(bs))
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
	privKey := ed25519.PrivateKey(pv.RawBytes())
	return SignatureFromRawBytes(ed25519.Sign(privKey, msg))

}

func (pv PrivateKey) PublicKey() PublicKey {
	publicKey, _ := PublicKeyFromRawBytes(pv.RawBytes()[32:])
	return publicKey
}
