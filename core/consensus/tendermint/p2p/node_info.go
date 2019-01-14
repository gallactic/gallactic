package p2p

import (
	net "github.com/tendermint/tendermint/p2p"
	amino "github.com/tendermint/go-amino"
)

const (
	maxNodeInfoSize = 10240 // 10Kb
	maxNumChannels  = 16    // plenty of room for upgrades, for now
)

// Max size of the NodeInfo struct
func MaxNodeInfoSize() int {
	return maxNodeInfoSize
}

// NodeInfo is the basic node information exchanged
// between two peers during the Tendermint P2P handshake.
type GNodeInfo struct {
	net.DefaultNodeInfo
}

//protobuf marshal,unmarshal and size
var cdc = amino.NewCodec()

func (info *GNodeInfo) Size() int {
	bs, _ := info.Marshal()
	return len(bs)
}

// Marshal returns the amino encoding.
func (info *GNodeInfo) Marshal() ([]byte, error) {
	return cdc.MarshalBinaryBare(info)
}

// MarshalTo calls Marshal and copies to the given buffer.
func (info *GNodeInfo) MarshalTo(data []byte) (int, error) {
	bs, err := info.Marshal()
	if err != nil {
		return -1, err
	}
	return copy(data, bs), nil
}

// Unmarshal deserializes from amino encoded form.
func (info *GNodeInfo) Unmarshal(bs []byte) error {
	return cdc.UnmarshalBinaryBare(bs, info)
}
