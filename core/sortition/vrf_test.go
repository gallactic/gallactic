package sortition

import (
	"testing"

	"github.com/gallactic/gallactic/crypto"
	"github.com/stretchr/testify/assert"
)

func TestVRF(t *testing.T) {
	for i := 0; i < 100; i++ {
		pk, pv := crypto.GenerateKey(nil)
		signer := crypto.NewValidatorSigner(pv)
		m := []byte{byte(i)}

		vrf := NewVRF(signer)

		max := uint64(i + 1*1000)
		vrf.SetMax(max)
		index, proof := vrf.Evaluate(m)

		//fmt.Printf("%x\n", index)
		assert.Equal(t, index <= max, true)

		index2, result := vrf.Verify(m, pk, proof)

		assert.Equal(t, result, true)
		assert.Equal(t, index, index2)
	}
}
