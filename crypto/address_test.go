package crypto

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
)

func TestAddress(t *testing.T) {
	bytes := []byte{
		1, 2, 3, 4, 5,
		1, 2, 3, 4, 5,
		1, 2, 3, 4, 5,
		1, 2, 3, 4, 5,
	}
	address1, err := addressFromHash(bytes, accountAddress)
	assert.NoError(t, err)
	word256 := address1.Word256()
	leadingZeroes := []byte{
		0, 0, 0, 0,
		0, 0,
	}
	fmt.Println(address1.String())
	assert.Equal(t, leadingZeroes, word256[:6])
	address2, err := AddressFromWord256(word256)
	assert.NoError(t, err)
	assert.Equal(t, address1, address2)
}

func TestMarshalingEmptyAddress(t *testing.T) {
	addr1 := Address{}

	js, err := json.Marshal(addr1)
	assert.NoError(t, err)
	assert.Equal(t, js, []byte("\"\""))
	var addr2 Address
	err = json.Unmarshal(js, &addr2)
	assert.Error(t, err)
	assert.Equal(t, addr1, addr2)

	bs, err := addr1.MarshalAmino()
	assert.NoError(t, err)
	assert.Equal(t, bs, []byte(nil))
	var addr3 Address
	err = json.Unmarshal(bs, &addr3)
	assert.Error(t, err)
	assert.Equal(t, addr1, addr3)
}

func TestMarshalingAddress(t *testing.T) {
	addrs := []string{
		"0123456789ABCDEF0123456789ABCDEF01234567",
		"7777777777777777777777777777777777777777",
		"B03DD2C47852775208A56FA10A49875ABC507343",
		"FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF",
	}

	for _, addr := range addrs {
		bs, _ := hex.DecodeString(addr)
		ac1, err := addressFromHash(bs, accountAddress)
		assert.NoError(t, err)
		va1, err := addressFromHash(bs, validatorAddress)
		assert.NoError(t, err)
		ct1, err := addressFromHash(bs, contractAddress)
		assert.NoError(t, err)
		fmt.Println(ac1.String())
		fmt.Println(va1.String())
		fmt.Println(ct1.String())

		jac, err := json.Marshal(&ac1)
		assert.NoError(t, err)
		jva, err := json.Marshal(&va1)
		assert.NoError(t, err)
		jct, err := json.Marshal(&ct1)
		assert.NoError(t, err)
		fmt.Println(string(jac))
		fmt.Println(string(jva))
		fmt.Println(string(jct))

		var ac2, va2, ct2 Address
		assert.NoError(t, json.Unmarshal(jac, &ac2))
		assert.NoError(t, json.Unmarshal(jva, &va2))
		assert.NoError(t, json.Unmarshal(jct, &ct2))

		require.Equal(t, ac1, ac2)
		require.Equal(t, va1, va2)
		require.Equal(t, ct1, ct2)

		bac, err := ac1.MarshalAmino()
		assert.NoError(t, err)
		bva, err := va1.MarshalAmino()
		assert.NoError(t, err)
		bct, err := ct1.MarshalAmino()
		assert.NoError(t, err)
		fmt.Println(string(jac))
		fmt.Println(string(jva))
		fmt.Println(string(jct))

		var ac3, va3, ct3 Address
		assert.NoError(t, ac3.UnmarshalAmino(bac))
		assert.NoError(t, va3.UnmarshalAmino(bva))
		assert.NoError(t, ct3.UnmarshalAmino(bct))

		require.Equal(t, ac1, ac2)
		require.Equal(t, va1, va2)
		require.Equal(t, ct1, ct2)
	}
}

func TestValidity(t *testing.T) {
	var err error
	_, err = AddressFromString("ac9E2cyNA5UfB8pUpqzEz4QCcBpp8sxnEaN")
	assert.NoError(t, err)

	_, err = AddressFromString("ac9E2cyNA5UfB8pUpqzEz4QCcBpp8sxnEaM")
	assert.Error(t, err)

	_, err = AddressFromString("009E2cyNA5UfB8pUpqzEz4QCcBpp8sxnEaM")
	assert.Error(t, err)

	_, err = AddressFromString("invalid_addres")
	assert.Error(t, err)

	_, err = AddressFromRawBytes([]byte{0, 1, 2, 3, 4, 5, 6})
	assert.Error(t, err)

	_, err = AddressFromRawBytes([]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5})
	assert.Error(t, err)
}
