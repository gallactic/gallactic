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
	var addr Address
	if err := addr.UnmarshalText([]byte(text)); err != nil {
		return Address{}, err
	}

	return addr, nil
}

func AddressFromRawByes(bs []byte) (Address, error) {
	var addr Address
	if err := addr.UnmarshalAmino(bs); err != nil {
		return Address{}, err
	}

	return addr, nil
}

func AddressFromWord256(w binary.Word256) (Address, error) {
	bs := w.Bytes()[6:]
	return AddressFromRawByes(bs)
}

/// this is a private constructor
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
	if err := addr.EnsureValid(); err != nil {
		return Address{}, err
	}

	return addr, nil
}

/// -------
/// CASTING

func (addr Address) RawBytes() []byte {
	return addr.data.Address[:]
}

func (addr Address) String() string {
	return base58.Encode(addr.data.Address[:])
}

func (addr Address) Word256() binary.Word256 {
	return binary.LeftPadWord256(addr.data.Address[:])
}

/// ----------
/// MARSHALING

func (addr Address) MarshalAmino() ([]byte, error) {
	return addr.data.Address[:], nil
}

func (addr *Address) UnmarshalAmino(bs []byte) error {
	if len(bs) != 26 {
		return e.Errorf(e.ErrInvalidAddress, "Address raw bytes should be 26 bytes, but it is %v bytes", len(bs))
	}

	copy(addr.data.Address[:], bs[:])
	if err := addr.EnsureValid(); err != nil {
		return err
	}

	return nil
}

func (addr Address) MarshalText() ([]byte, error) {
	return []byte(addr.String()), nil
}

func (addr *Address) UnmarshalText(text []byte) error {
	bs, err := base58.Decode(string(text))
	if err != nil {
		return err
	}

	return addr.UnmarshalAmino(bs)
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

func (addr *Address) IsValidatorAddress() bool {
	return addr.version() == validatorAddress
}

func (addr *Address) IsAccountAddress() bool {
	return addr.version() == accountAddress
}

func (addr Address) EqualsTo(right Address) bool {
	return bytes.Equal(addr.RawBytes(), right.RawBytes())
}
