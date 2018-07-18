package account

import (
	"fmt"
	"strconv"
	"strings"
)

type Permissions uint64

func (p *Permissions) Set(r Permissions) {
	*p |= r
}

func (p *Permissions) Unset(r Permissions) {
	*p &= ^r
}

func (p Permissions) IsSet(r Permissions) bool {
	return (p & r) == r
}

func (p Permissions) String() string {
	return fmt.Sprintf("%#v", p)
}

func (p Permissions) MarshalText() ([]byte, error) {
	str := p.String()
	return []byte(str), nil
}

func (p *Permissions) UnmarshalJSON(bs []byte) error {
	cleaned := string(bs)
	cleaned = strings.Replace(cleaned, "0x", "", -1)
	cleaned = strings.Replace(cleaned, "\"", "", -1)

	val, err := strconv.ParseUint(cleaned, 16, 64)
	if err != nil {
		return err
	}
	*p = Permissions(val)
	return nil
}
