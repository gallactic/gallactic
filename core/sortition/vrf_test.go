package sortition

import (
	"testing"
)

func TestVRF(t *testing.T) {
	for i := 0; i < 100; i++ {
		/*
			pv, _ := acm.GeneratePrivateKey(nil)
			pa, _ := acm.GeneratePrivateAccountFromPrivateKeyBytes(pv.Bytes()[1:])
			pk := pv.PublicKey()
			m := []byte{byte(i)}

			vrf := finterra.NewVRF(pa, pa)

			var max uint64 = uint64(i + 1*1000)
			vrf.SetMax(max)
			index, proof := vrf.Evaluate(m)

			//fmt.Printf("%x\n", index)
			assert.Equal(t, index <= max, true)

			index2, result := vrf.Verify(m, pk, proof)

			assert.Equal(t, result, true)
			assert.Equal(t, index, index2)
		*/
	}
}
