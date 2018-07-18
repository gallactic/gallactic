package crypto

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
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

func PrivateKeyFromRawBytes(bs []byte) (PrivateKey, error) {
	pv := PrivateKey{
		data: privateKeyData{
			PrivateKey: bs,
		},
	}

	if err := pv.check(); err != nil {
		return PrivateKey{}, err
	}

	return pv, nil
}

func PrivateKeyFromSecret(secret string) PrivateKey {
	hasher := sha256.New()
	hasher.Write(([]byte)(secret))
	// No error from a buffer
	privateKey, _ := GeneratePrivateKey(bytes.NewBuffer(hasher.Sum(nil)))
	return privateKey
}

func GeneratePrivateKey(random io.Reader) (PrivateKey, error) {
	if random == nil {
		random = rand.Reader
	}
	_, privKey, err := ed25519.GenerateKey(random)
	if err != nil {
		return PrivateKey{}, err
	}

	return PrivateKeyFromRawBytes(privKey)
}

func (pv PrivateKey) check() error {
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

/// -------
/// CASTING

func (pv *PrivateKey) IsValid() bool {
	return pv.check() == nil
}

func (pv PrivateKey) RawBytes() []byte {
	return pv.data.PrivateKey
}

func (pv PrivateKey) String() string {
	return hex.EncodeToString(pv.RawBytes())
}

/// ----------
/// MARSHALING

func (pv PrivateKey) MarshalText() ([]byte, error) {
	return json.Marshal(pv.data)
}

func (pv *PrivateKey) UnmarshalText(bs []byte) error {
	str := string(bs)
	bs, err := hex.DecodeString(str)
	if err != nil {
		return err
	}

	p, err := PrivateKeyFromRawBytes(bs)
	if err != nil {
		return err
	}

	*pv = p
	return nil
}

/// ----------
/// ATTRIBUTES

func (pv PrivateKey) Sign(msg []byte) (Signature, error) {
	privKey := ed25519.PrivateKey(pv.data.PrivateKey)
	return SignatureFromRawBytes(ed25519.Sign(privKey, msg))

}

func (pv PrivateKey) PublicKey() PublicKey {
	publicKey, _ := PublicKeyFromRawBytes(pv.RawBytes()[32:])
	return publicKey
}
