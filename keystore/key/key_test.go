package key

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncryption(t *testing.T) {
	auth := "1234"
	//Generates Private Key
	k1 := GenAccountKey()

	fname := fmt.Sprintf("%s.json", k1.Address().String())

	//Encrypts Private Key
	data, err := EncryptKeyFile(k1, auth)
	assert.NoError(t, err)

	err = storeNewKey(k1, auth, fname)
	assert.NoError(t, err)

	kj := new(encryptedKeyJSONV3)
	if err := json.Unmarshal(data, kj); err != nil {
		assert.NoError(t, err)
	}

	//Decrypts Json Object
	k2, err := DecryptKeyFile(filePath+fname, auth)
	assert.NoError(t, err)

	if !reflect.DeepEqual(k1.data.PrivateKey, k2.data.PrivateKey) {
		t.Fatal(err)
	}
}
