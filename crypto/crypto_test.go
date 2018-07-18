package crypto

import (
	"fmt"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGlobalAddress(t *testing.T) {
	addr := "0000000000000000000000000000000000000000"
	bs, _ := hex.DecodeString(addr)
	gb, err := addressFromHash(bs, globalAddress)
	fmt.Println(gb.String())

	assert.NoError(t, err)
	assert.Equal(t, GlobalAddress, gb)
}
