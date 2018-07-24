package crypto

import (
	"github.com/gallactic/gallactic/common/binary"
	"golang.org/x/crypto/ripemd160"
)

var GlobalAddress, _ = AddressFromString("gbnDyGpPVijzB74qUhiEAvgdZYMdK2Uvdkh")

func DeriveContractAddress(addr Address, sequence uint64) Address {
	temp := make([]byte, 32+8)
	copy(temp, addr.data.Address[:])
	binary.PutUint64BE(temp[32:], uint64(sequence))
	hasher := ripemd160.New()
	hasher.Write(temp) // does not error
	hash := hasher.Sum(nil)

	ct, _ := ContractAddress(hash)
	return ct
}
