package evm

import (
	"encoding/json"
	"testing"

	"github.com/gallactic/gallactic/common/binary"
	"github.com/gallactic/gallactic/crypto"
	"github.com/stretchr/testify/assert"
)

func TestMarshaling(t *testing.T) {
	log1 := Logs{{Address: crypto.GlobalAddress, Topics: []binary.HexBytes{{1, 2}, {3}}, Data: binary.HexBytes{1, 2}}}

	bs, err := log1.MarshalBinary()
	assert.NoError(t, err)
	var log2 Logs
	err = log2.UnmarshalBinary(bs)
	assert.NoError(t, err) /// No error
	assert.Equal(t, log1, log2)

	jsonLog, err := json.Marshal(log1)
	assert.NoError(t, err)
	var log3 Logs
	err = json.Unmarshal(jsonLog, &log3)
	assert.NoError(t, err) /// No error
	assert.Equal(t, log1, log3)
}
