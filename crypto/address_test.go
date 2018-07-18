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
	}
}

/*
Ankur:Please check current tests and re-write them

func TestNewContractAddress(t *testing.T) {
	addr := NewContractAddress(Address{
		233, 181, 216, 115, 19,
		53, 100, 101, 250, 227,
		60, 64, 108, 226, 194,
		151, 157, 230, 11, 203,
	}, 1)

	assert.Equal(t, Address{
		73, 234, 48, 252, 174,
		115, 27, 222, 54, 116,
		47, 133, 144, 21, 73,
		245, 21, 234, 26, 50,
	}, addr)
}

func TestAddress_MarshalJSON(t *testing.T) {
	addr := Address{
		73, 234, 48, 252, 174,
		115, 27, 222, 54, 116,
		47, 133, 144, 21, 73,
		245, 21, 234, 26, 50,
	}

	bs, err := json.Marshal(addr)
	assert.NoError(t, err)

	addrOut := new(Address)
	err = json.Unmarshal(bs, addrOut)

	assert.Equal(t, addr, *addrOut)
}

func TestAddress_MarshalText(t *testing.T) {
	addr := Address{
		73, 234, 48, 252, 174,
		115, 27, 222, 54, 116,
		47, 133, 144, 21, 73,
		245, 21, 234, 26, 50,
	}

	bs, err := addr.MarshalText()
	assert.NoError(t, err)

	addrOut := new(Address)
	err = addrOut.UnmarshalText(bs)

	assert.Equal(t, addr, *addrOut)
}

func TestAddress_Length(t *testing.T) {
	addrOut := new(Address)
	err := addrOut.UnmarshalText(([]byte)("49EA30FCAE731BDE36742F85901549F515EA1A10"))
	require.NoError(t, err)

	err = addrOut.UnmarshalText(([]byte)("49EA30FCAE731BDE36742F85901549F515EA1A1"))
	assert.Error(t, err, "address too short")

	err = addrOut.UnmarshalText(([]byte)("49EA30FCAE731BDE36742F85901549F515EA1A1020"))
	assert.Error(t, err, "address too long")
}

func TestAddress_Sort(t *testing.T) {
	addresses := Addresses{
		{2, 3, 4},
		{3, 1, 2},
		{2, 1, 2},
	}
	sorted := make(Addresses, len(addresses))
	copy(sorted, addresses)
	sort.Stable(sorted)
	assert.Equal(t, Addresses{
		{2, 1, 2},
		{2, 3, 4},
		{3, 1, 2},
	}, sorted)
}
*/
