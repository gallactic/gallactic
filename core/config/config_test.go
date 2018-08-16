package config

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoading(t *testing.T) {
	conf1 := defaultConfig()
	conf1.Tendermint.ListenAddress = "1.1.1.1"
	conf1.Tendermint.Moniker = "moniker"
	conf1.Tendermint.TendermintRoot = "tendermint"
	s := conf1.ToTOML()
	fmt.Println(s)
	conf2, err := FromTOML(s)
	require.NoError(t, err)
	require.Equal(t, conf1, conf2)
}
