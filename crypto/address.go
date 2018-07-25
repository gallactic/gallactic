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
	globalAddress    uint16 = 0x4C16
)

type Address struct {
	data addressData
}

type addressData struct {
	Address [26]byte
}

// checksum: first four bytes of sha256^2
func checksum(input []byte) (chksum [4]byte) {
	h := sha256.Sum256(input)
	h2 := sha256.Sum256(h[:])
	copy(chksum[:], h2[:4])
	return
}

/// ------------
/// CONSTRUCTORS

func AddressFromString(text string) (Address, error) {
	bs, err := base58.Decode(text)
	if err != nil {
		return Address{}, e.Errorf(e.ErrInvalidAddress, "%v", err.Error())
	}

	return AddressFromRawBytes(bs)
}

func AddressFromRawBytes(bs []byte) (Address, error) {
	if len(bs) != 26 {
		return Address{}, e.Errorf(e.ErrInvalidAddress, "Address raw bytes should be 26 bytes, but it is %v bytes", len(bs))
	}

	var addr Address
	copy(addr.data.Address[:], bs[:])
	if err := addr.EnsureValid(); err != nil {
		return Address{}, err
	}

	return addr, nil
}

func AddressFromWord256(w binary.Word256) (Address, error) {
	bs := w.Bytes()[6:]
	return AddressFromRawBytes(bs)
}

func ContractAddress(bs []byte) (Address, error) {
	return addressFromHash(bs, contractAddress)
}

func AccountAddress(bs []byte) (Address, error) {
	return addressFromHash(bs, accountAddress)
}

/// this is a private constructor
func addressFromHash(hash []byte, ver uint16) (Address, error) {
	if len(hash) != 20 {
		return Address{}, e.Errorf(e.ErrInvalidAddress, "Address hash should be 20 bytes but it is %v bytes", len(hash))
	}

	bs := make([]byte, 0, 2+20+4)
	bs = append(bs, (*[2]byte)(unsafe.Pointer(&ver))[:]...)
	bs = append(bs, hash...)
	chksum := checksum(bs)
	bs = append(bs, chksum[:]...)

	var addr Address
	copy(addr.data.Address[:], bs[:])
	if err := addr.EnsureValid(); err != nil {
		return Address{}, err
	}

	return addr, nil
}

/// -------
/// CASTING

func (addr Address) RawBytes() []byte {
	if addr.data.Address == [26]byte{} {
		return nil
	}

	return addr.data.Address[:]
}

func (addr Address) String() string {
	return base58.Encode(addr.RawBytes())
}

func (addr Address) Word256() binary.Word256 {
	return binary.LeftPadWord256(addr.RawBytes())
}

/// ----------
/// MARSHALING

func (addr Address) MarshalAmino() ([]byte, error) {
	return addr.RawBytes(), nil
}

func (addr *Address) UnmarshalAmino(bs []byte) error {
	/// when the address is empty, unmarshal it as empty address
	if len(bs) == 0 {
		return nil
	}

	a, err := AddressFromRawBytes(bs)
	if err != nil {
		return err
	}

	*addr = a
	return nil
}

func (addr Address) MarshalText() ([]byte, error) {
	return []byte(addr.String()), nil
}

func (addr *Address) UnmarshalText(text []byte) error {
	/// when the address is empty, unmarshal it as empty address
	if len(text) == 0 {
		return nil
	}

	a, err := AddressFromString(string(text))
	if err != nil {
		return err
	}

	*addr = a
	return nil
}

/// -------
/// METHODS

func (addr *Address) EnsureValid() error {
	bs := addr.RawBytes()
	chksum1 := addr.checksum()
	chksum2 := checksum(bs[0:22])
	if chksum1 != chksum2 {
		return e.Errorf(e.ErrInvalidAddress, "Checksum doesn't match. It should be %v but it is %v", chksum1, chksum2)
	}

	ver := (uint16(bs[1])<<8 | uint16(bs[0]))
	if ver != accountAddress && ver != validatorAddress &&
		ver != contractAddress && ver != globalAddress {
		return e.Errorf(e.ErrInvalidAddress, "Invalid version: %X", ver)
	}

	return nil
}

func (addr Address) Verify(pb PublicKey) bool {
	if addr.IsAccountAddress() {
		return pb.AccountAddress().EqualsTo(addr)
	} else if addr.IsValidatorAddress() {
		return pb.ValidatorAddress().EqualsTo(addr)
	}

	return false
}

func (addr Address) version() uint16 {
	bs := addr.RawBytes()
	return (uint16(bs[1])<<8 | uint16(bs[0]))
}

func (addr Address) checksum() (chksum [4]byte) {
	bs := addr.RawBytes()
	copy(chksum[:], bs[22:26])
	return
}

func (addr *Address) IsContractAddress() bool {
	return addr.version() == contractAddress
}

func (addr *Address) IsValidatorAddress() bool {
	return addr.version() == validatorAddress
}

func (addr *Address) IsAccountAddress() bool {
	if addr.version() == accountAddress {
		return true
	}

	return addr.EqualsTo(GlobalAddress) /// Global address technically is an account address
}

func (addr Address) EqualsTo(right Address) bool {
	return bytes.Equal(addr.RawBytes(), right.RawBytes())
}
