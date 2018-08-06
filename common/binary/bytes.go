package binary

import "encoding/hex"

type HexBytes []byte

func (hb *HexBytes) UnmarshalText(hexBytes []byte) error {
	bs, err := hex.DecodeString(string(hexBytes))
	if err != nil {
		return err
	}
	*hb = bs
	return nil
}

func (hb HexBytes) MarshalText() ([]byte, error) {
	return []byte(hb.String()), nil
}

func (hb HexBytes) String() string {
	return hex.EncodeToString(hb)
}
