package key

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncryption(t *testing.T) {
	k1 := GenAccountKey()

	fname := fmt.Sprintf("%s.json", k1.Address().String())
	err := EncryptKeyFile(k1, fname, "1234")
	assert.NoError(t, err)
	k2, err := DecryptKeyFile(fname, "1234")
	assert.NoError(t, err)
	assert.Equal(t, k1, k2)
}
