package key

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	crand "crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/big"
	"os"

	"github.com/gallactic/gallactic/core/evm/sha3"
	"github.com/gallactic/gallactic/crypto"
	"golang.org/x/crypto/pbkdf2"
	"golang.org/x/crypto/scrypt"
)

const (
	keyHeaderKDF = "scrypt"
	scryptN      = 2
	scryptP      = 1
	scryptR      = 8
	scryptDKLen  = 32

	// number of bits in a big.Word
	wordBits = 32 << (uint64(^big.Word(0)) >> 63)
	// number of bytes in a big.Word
	wordBytes = wordBits / 8
	version   = 3

	filePath = "/tmp/"
)

type encryptedKeyJSONV3 struct {
	Crypto  cryptoJSON `json:"crypto"`
	Version int        `json:"version"`
}

type cryptoJSON struct {
	Cipher       string                 `json:"cipher"`
	CipherText   string                 `json:"ciphertext"`
	CipherParams cipherparamsJSON       `json:"cipherparams"`
	KDF          string                 `json:"kdf"`
	KDFParams    map[string]interface{} `json:"kdfparams"`
	MAC          string                 `json:"mac"`
}

type cipherparamsJSON struct {
	IV string `json:"iv"`
}

// DecryptKeyFile returns an instance of Key object
func DecryptKeyFile(file, auth string) (*Key, error) {
	data, err := ioutil.ReadFile(file)

	kj := new(encryptedKeyJSONV3)
	if err := json.Unmarshal(data, kj); err != nil {
		return nil, err
	}
	keyBytes, err := DecryptKey(kj, auth)
	if err != nil {
		return nil, err
	}
	pv, err := crypto.PrivateKeyFromRawBytes(keyBytes)
	return &Key{
		data: keyData{
			PrivateKey: pv,
			PublicKey:  pv.PublicKey(),
			Address:    pv.PublicKey().AccountAddress(),
		},
	}, nil
}

func DecryptKey(kj *encryptedKeyJSONV3, auth string) (keyBytes []byte, err error) {

	if kj.Crypto.Cipher != "aes-128-ctr" {
		return nil, fmt.Errorf("Cipher not supported: %v", kj.Crypto.Cipher)
	}

	mac, err := hex.DecodeString(kj.Crypto.MAC)
	if err != nil {
		return nil, err
	}

	iv, err := hex.DecodeString(kj.Crypto.CipherParams.IV)
	if err != nil {
		return nil, err
	}

	cipherText, err := hex.DecodeString(kj.Crypto.CipherText)
	if err != nil {
		return nil, err
	}

	derivedKey, err := getKDFKey(kj.Crypto, auth)
	if err != nil {
		return nil, err
	}

	calculatedMAC := sha3.Sha3(derivedKey[16:32], cipherText)
	if !bytes.Equal(calculatedMAC, mac) {
		return nil, err
	}

	plainText, err := aesCTRXOR(derivedKey[:16], cipherText, iv)
	if err != nil {
		return nil, err
	}

	return plainText, err

}

func getKDFKey(cryptoJSON cryptoJSON, auth string) ([]byte, error) {

	authArray := []byte(auth)
	salt, err := hex.DecodeString(cryptoJSON.KDFParams["salt"].(string))
	if err != nil {
		return nil, err
	}
	dkLen := ensureInt(cryptoJSON.KDFParams["dklen"])

	if cryptoJSON.KDF == keyHeaderKDF {
		n := ensureInt(cryptoJSON.KDFParams["n"])
		r := ensureInt(cryptoJSON.KDFParams["r"])
		p := ensureInt(cryptoJSON.KDFParams["p"])
		return scrypt.Key(authArray, salt, n, r, p, dkLen)

	} else if cryptoJSON.KDF == "pbkdf2" {
		c := ensureInt(cryptoJSON.KDFParams["c"])
		prf := cryptoJSON.KDFParams["prf"].(string)
		if prf != "hmac-sha256" {
			return nil, fmt.Errorf("Unsupported PBKDF2 PRF: %s", prf)
		}
		key := pbkdf2.Key(authArray, salt, c, dkLen, sha256.New)
		return key, nil
	}

	return nil, fmt.Errorf("Unsupported KDF: %s", cryptoJSON.KDF)
}

// EncryptKeyFile encrypts a key and return encrypted byte array
func EncryptKeyFile(key *Key, auth string) ([]byte, error) {
	authArray := []byte(auth)
	salt := GetEntropyCSPRNG(32)
	derivedKey, err := scrypt.Key(authArray, salt, scryptN, scryptR, scryptP, scryptDKLen)
	if err != nil {
		return nil, err
	}

	encryptKey := derivedKey[:16]
	keyBytes := key.PrivateKey().RawBytes()

	iv := GetEntropyCSPRNG(aes.BlockSize) // 16
	cipherText, err := aesCTRXOR(encryptKey, keyBytes, iv)
	if err != nil {
		return nil, err
	}
	mac := sha3.Sha3(derivedKey[16:32], cipherText)

	scryptParamsJSON := make(map[string]interface{}, 5)
	scryptParamsJSON["n"] = scryptN
	scryptParamsJSON["r"] = scryptR
	scryptParamsJSON["p"] = scryptP
	scryptParamsJSON["dklen"] = scryptDKLen
	scryptParamsJSON["salt"] = hex.EncodeToString(salt)

	cipherParamsJSON := cipherparamsJSON{
		IV: hex.EncodeToString(iv),
	}

	cryptoStruct := cryptoJSON{
		Cipher:       "aes-128-ctr",
		CipherText:   hex.EncodeToString(cipherText),
		CipherParams: cipherParamsJSON,
		KDF:          keyHeaderKDF,
		KDFParams:    scryptParamsJSON,
		MAC:          hex.EncodeToString(mac),
	}

	encryptedKeyJSONV3 := encryptedKeyJSONV3{
		cryptoStruct,
		version,
	}

	return json.Marshal(encryptedKeyJSONV3)
}

func storeNewKey(k *Key, auth, fname string) error {

	bs, err := EncryptKeyFile(k, auth)
	if err != nil {
		return err
	}

	fileName := filePath + fname
	error := writeKeyFile(fileName, bs)
	if error != nil {
		return error
	}
	return nil
}

func GetEntropyCSPRNG(n int) []byte {
	mainBuff := make([]byte, n)
	_, err := io.ReadFull(crand.Reader, mainBuff)
	if err != nil {
		panic("reading from crypto/rand failed: " + err.Error())
	}
	return mainBuff
}

func aesCTRXOR(key, inText, iv []byte) ([]byte, error) {
	// AES-128 is selected due to size of encryptKey.
	aesBlock, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	stream := cipher.NewCTR(aesBlock, iv)
	outText := make([]byte, len(inText))
	stream.XORKeyStream(outText, inText)
	return outText, err
}

// why do integers in KDF params end up as float64 and not int after
// unmarshal?
func ensureInt(x interface{}) int {
	res, ok := x.(int)
	if !ok {
		res = int(x.(float64))
	}
	return res
}

func writeKeyFile(filepath string, content []byte) error {

	f, err := os.Create(filepath)
	if err != nil {
		return err
	}
	if b, err := f.Write(content); err != nil {
		fmt.Printf("wrote %d bytes\n", b)
		f.Close()
		return err
	}

	f.Close()
	return os.Rename(f.Name(), filepath)
}
