package key

import (
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

	//Encrypts the key json blob
	err := EncryptKeyFile(k1, auth, fname)
	assert.NoError(t, err)

	//Decrypts Json Object
	k2, err := DecryptKeyFile(auth, fname)
	assert.NoError(t, err)

	if !reflect.DeepEqual(k1, k2) {
		t.Fatal(err)
	}
}
