package crypto

import (
	"bytes"
	"unsafe"

	e "github.com/gallactic/gallactic/errors"
	"github.com/mr-tron/base58/base58"
)

const AddressSize = 26

type Address struct {
	data addressData
}

type addressData struct {
	Address [AddressSize]byte
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
	if len(bs) != AddressSize {
		return Address{}, e.Errorf(e.ErrInvalidAddress, "Address should be %d bytes, but it is %v bytes", AddressSize, len(bs))
	}

	var addr Address
	copy(addr.data.Address[:], bs[:])
	if err := addr.EnsureValid(); err != nil {
		return Address{}, err
	}

	return addr, nil
}

func ContractAddress(bs []byte) (Address, error) {
	return addressFromHash(bs, prefixContractAddress)
}

func AccountAddress(bs []byte) (Address, error) {
	return addressFromHash(bs, prefixAccountAddress)
}

func ValidatorAddress(bs []byte) (Address, error) {
	return addressFromHash(bs, prefixValidatorAddress)
}

/// this is a private constructor
func addressFromHash(hash []byte, prefix uint16) (Address, error) {
	if len(hash) != 20 {
		return Address{}, e.Errorf(e.ErrInvalidAddress, "Address hash should be 20 bytes but it is %v bytes", len(hash))
	}

	data := make([]byte, 0, AddressSize)
	data = append(data, (*[2]byte)(unsafe.Pointer(&prefix))[:]...)
	data = append(data, hash...)
	chksum := checksum(data)
	data = append(data, chksum[:]...)

	return AddressFromRawBytes(data)
}

/// -------
/// CASTING

func (addr Address) RawBytes() []byte {
	if addr.data.Address == [AddressSize]byte{} {
		return nil
	}

	return addr.data.Address[:]
}

func (addr Address) String() string {
	return base58.Encode(addr.RawBytes())
}

/// ----------
/// MARSHALING

func (addr Address) MarshalAmino() ([]byte, error) {
	return addr.RawBytes(), nil
}

func (addr *Address) UnmarshalAmino(bs []byte) error {
	/// Unmarshal empty value
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
	/// Unmarshal empty value
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

// Gogo proto support
func (addr *Address) Marshal() ([]byte, error) {
	return addr.MarshalAmino()
}

func (addr *Address) Unmarshal(bs []byte) error {
	return addr.UnmarshalAmino(bs)
}

func (addr *Address) MarshalTo(data []byte) (int, error) {
	return copy(data, addr.data.Address[:]), nil
}

func (addr *Address) Size() int {
	return AddressSize
}

/// -------
/// METHODS

func (addr *Address) EnsureValid() error {
	bs := addr.RawBytes()
	err := validateChecksum(bs)
	if err != nil {
		return e.Errorf(e.ErrInvalidAddress, err.Error())
	}

	err = validatePrefix(bs, prefixAccountAddress, prefixValidatorAddress, prefixContractAddress, prefixGlobalAddress)
	if err != nil {
		return e.Errorf(e.ErrInvalidAddress, err.Error())
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

func (addr Address) prefix() uint16 {
	bs := addr.RawBytes()
	return (uint16(bs[1])<<8 | uint16(bs[0]))
}
func (addr *Address) IsContractAddress() bool {
	return addr.prefix() == prefixContractAddress
}

func (addr *Address) IsValidatorAddress() bool {
	return addr.prefix() == prefixValidatorAddress
}

func (addr *Address) IsAccountAddress() bool {
	if addr.prefix() == prefixAccountAddress {
		return true
	}

	return addr.EqualsTo(GlobalAddress) /// Global address technically is an account address
}

func (addr Address) EqualsTo(right Address) bool {
	return bytes.Equal(addr.RawBytes(), right.RawBytes())
}
