package crypto

import (
	"crypto/sha256"
	"fmt"

	"github.com/gallactic/gallactic/common/binary"
	"golang.org/x/crypto/ripemd160"
)

const (
	prefixAccountAddress   uint16 = 0xEC12 // ac..
	prefixValidatorAddress uint16 = 0x2A1E // va..
	prefixContractAddress  uint16 = 0x3414 // ct..
	prefixGlobalAddress    uint16 = 0x4C16 // gb..
	prefixPublicKey        uint16 = 0x9005 // pj,pk,pm..
	prefixPrivateKey       uint16 = 0xE913 // sk..
)

// checksum: first four bytes of sha256^2
func checksum(input []byte) (chksum [4]byte) {
	h := sha256.Sum256(input)
	h2 := sha256.Sum256(h[:])
	copy(chksum[:], h2[:4])
	return
}

func validateChecksum(bs []byte) error {
	l := len(bs)
	var chksum1 [4]byte
	copy(chksum1[:], bs[l-4:])
	chksum2 := checksum(bs[0 : l-4])
	if chksum1 != chksum2 {
		return fmt.Errorf("Checksum doesn't match. Expected %v, got %v", chksum1, chksum2)
	}
	return nil
}

func validatePrefix(bs []byte, prefixes ...uint16) error {
	pre := (uint16(bs[1])<<8 | uint16(bs[0]))
	for _, p := range prefixes {
		if p == pre {
			return nil
		}
	}
	return fmt.Errorf("Invalid prefix: %X", pre)
}

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
