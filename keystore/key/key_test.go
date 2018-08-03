package key

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncryption(t *testing.T) {
	auth := "secret"
	//Generates Private Key
	k1 := GenAccountKey()
	filePath := fmt.Sprintf("/tmp/%s.key", k1.Address().String())
	//Encrypts the key json blob
	err := EncryptKeyFile(k1, filePath, auth)
	assert.NoError(t, err)
	//Decrypts Json Object
	k2, err := DecryptKeyFile(filePath, auth)
	assert.NoError(t, err)
	assert.Equal(t, k1, k2)
	// wrong password: should fails
	k3, err := DecryptKeyFile(filePath, "Secret")
	assert.Error(t, err)
	assert.Nil(t, k3)
	// invalid file path, should fails
	filePath1 := fmt.Sprintf("/tmp/%s_invalid_path.key", k1.Address().String())
	k4, err := DecryptKeyFile(filePath1, auth)
	fmt.Println(err)
	assert.Error(t, err)
	assert.Nil(t, k4)
}

func TestEncryptionData(t *testing.T) {
	auth := "secret"
	//Generates
	k1 := GenValidatorKey()
	k2 := GenAccountKey()
	k3, err := NewKey(k1.Address(), k2.PrivateKey())

	assert.NotNil(t, k1)
	assert.NotNil(t, k2)
	assert.Nil(t, k3)
	assert.Error(t, err)
}
