package crypto

import (
	"bytes"
	"crypto/sha256"
	"unsafe"

	"github.com/gallactic/gallactic/common/binary"
	"github.com/gallactic/gallactic/errors"
	"github.com/mr-tron/base58/base58"
)

const (
	accountAddress   uint16 = 0xEC12 // ac...
	validatorAddress uint16 = 0x2A1E //
	contractAddress  uint16 = 0x3414
	globalAddress    uint16 = 0x4A16
)

type Address struct {
	data addressData
}

type addressData struct {
	Address [26]byte
}

// checksum: first four bytes of sha256^2
func checksum(input []byte) (cksum [4]byte) {
	h := sha256.Sum256(input)
	h2 := sha256.Sum256(h[:])
	copy(cksum[:], h2[:4])
	return
}

/// ------------
/// CONSTRUCTORS

func AddressFromString(s string) (Address, error) {
	var addr Address
	bs, err := base58.Decode(s)
	if err != nil {
		return addr, err
	}

	return AddressFromRawByes(bs)
}

func AddressFromRawByes(bs []byte) (Address, error) {
	var addr Address

	copy(addr.data.Address[:], bs[:])
	if err := addr.check(); err != nil {
		return Address{}, err
	}

	return addr, nil
}

/// this is private constructor
func addressFromHash(hash []byte, ver uint16) (Address, error) {
	var addr Address
	if len(hash) != 20 {
		return addr, e.Errorf(e.ErrInvalidAddress, "Address hash should be 20 bytes but it is %v bytes", len(hash))
	}

	bs := make([]byte, 0, 2+20+4)
	bs = append(bs, (*[2]byte)(unsafe.Pointer(&ver))[:]...)
	bs = append(bs, hash...)
	chksum := checksum(bs)
	bs = append(bs, chksum[:]...)

	copy(addr.data.Address[:], bs[:])
	if err := addr.check(); err != nil {
		return Address{}, err
	}

	return addr, nil
}

func AddressFromWord256(w binary.Word256) (Address, error) {
	bs := w.Bytes()[6:]
	return AddressFromRawByes(bs)
}

func (addr *Address) check() error {
	bs := addr.RawBytes()
	chksum1 := bs[22:26]
	chksum2 := checksum(bs[0:22])
	if !bytes.Equal(chksum1, chksum2[:]) {
		return e.Errorf(e.ErrInvalidAddress, "Checksum doesn't match. It should be %v but it is %v", chksum1, chksum2)
	}

	ver := (uint16(bs[1])<<8 | uint16(bs[0]))
	if ver != accountAddress && ver != validatorAddress &&
		ver != contractAddress && ver != globalAddress {
		return e.Errorf(e.ErrInvalidAddress, "Invalid version: %X", ver)
	}

	return nil
}

/// -------
/// CASTING

func (addr Address) Word256() binary.Word256 {
	return binary.LeftPadWord256(addr.data.Address[:])
}

func (addr Address) RawBytes() []byte {
	return addr.data.Address[:]
}

func (addr Address) String() string {
	return base58.Encode(addr.data.Address[:])
}

/// ----------
/// MARSHALING

func (addr Address) MarshalText() ([]byte, error) {
	return []byte(addr.String()), nil
}

func (addr *Address) UnmarshalText(bs []byte) error {
	str := string(bs)
	a, err := AddressFromString(str)
	if err != nil {
		return err
	}

	*addr = a
	return nil
}

/// ----------
/// ATTRIBUTES

func (addr *Address) IsValid() bool {
	return addr.check() == nil
}

func (addr *Address) IsValidatorAddress() bool {
	return false
}

func (addr *Address) IsAccountAddress() bool {
	return false
}

func (addr Address) EqualsTo(right Address) bool {
	return bytes.Equal(addr.RawBytes(), right.RawBytes())
}
