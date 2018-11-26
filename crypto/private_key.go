package crypto

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"io"
	"unsafe"

	"github.com/gallactic/gallactic/errors"
	"github.com/mr-tron/base58/base58"
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

// PrivateKeyFromString constructs a private key from base58 encoding text and check the prefix and checksum
func PrivateKeyFromString(text string) (PrivateKey, error) {
	data, err := base58.Decode(text)
	if err != nil {
		return PrivateKey{}, e.Errorf(e.ErrInvalidPrivateKey, "%v", err.Error())
	}

	if len(data) != ed25519.PrivateKeySize+6 {
		return PrivateKey{}, e.Errorf(e.ErrInvalidPrivateKey, "Private key should be %v bytes but it is %v bytes", ed25519.PrivateKeySize+6, len(data))
	}

	err = validateChecksum(data)
	if err != nil {
		return PrivateKey{}, e.Errorf(e.ErrInvalidPrivateKey, err.Error())
	}

	err = validatePrefix(data, prefixPrivateKey)
	if err != nil {
		return PrivateKey{}, e.Errorf(e.ErrInvalidPrivateKey, err.Error())
	}

	bs := data[2 : ed25519.PrivateKeySize+2]
	return PrivateKeyFromRawBytes(bs)
}

// PrivateKeyFromRawBytes constructs a private key from ed25519 raw bytes.
func PrivateKeyFromRawBytes(bs []byte) (PrivateKey, error) {
	/// Check for empty value
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

func GenerateKeyFromSecret(secret string) (PublicKey, PrivateKey) {
	hasher := sha256.New()
	hasher.Write(([]byte)(secret))

	return GenerateKey(bytes.NewBuffer(hasher.Sum(nil)))
}

func GenerateKey(random io.Reader) (PublicKey, PrivateKey) {
	if random == nil {
		random = rand.Reader
	}
	// No error from a buffer
	pubKey, privKey, _ := ed25519.GenerateKey(random)
	pk, _ := PublicKeyFromRawBytes(pubKey)
	pv, _ := PrivateKeyFromRawBytes(privKey)
	return pk, pv
}

/// -------
/// CASTING
// RawBytes returns the ed25519 raw bytes of the private key
func (pv PrivateKey) RawBytes() []byte {
	return pv.data.PrivateKey
}

// String return the base58 encoding text of the private key with the prefix and checksum
func (pv PrivateKey) String() string {
	if len(pv.data.PrivateKey) == 0 {
		return ""
	}

	prefix := prefixPrivateKey
	data := make([]byte, 0, ed25519.PrivateKeySize+6)
	data = append(data, (*[2]byte)(unsafe.Pointer(&prefix))[:]...)
	data = append(data, pv.data.PrivateKey...)
	chksum := checksum(data)
	data = append(data, chksum[:]...)

	return base58.Encode(data)
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
	/// Unmarshal empty text
	if len(text) == 0 {
		return nil
	}

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
	return SignatureFromRawBytes(ed25519.Sign(privKey, Sha3(msg)))
}

func (pv PrivateKey) SignWithoutHash(msg []byte) (Signature, error) {
	privKey := ed25519.PrivateKey(pv.RawBytes())
	return SignatureFromRawBytes(ed25519.Sign(privKey, msg))
}

func (pv PrivateKey) PublicKey() PublicKey {
	publicKey, _ := PublicKeyFromRawBytes(pv.RawBytes()[32:])
	return publicKey
}
