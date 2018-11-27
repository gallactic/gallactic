package config

import (
	"bytes"
	"io/ioutil"

	"github.com/BurntSushi/toml"
	"github.com/gallactic/gallactic/common"
	tmConfig "github.com/gallactic/gallactic/core/consensus/tendermint/config"
	rpcConfig "github.com/gallactic/gallactic/rpc/config"
	grpcConfig "github.com/gallactic/gallactic/rpc/grpc/config"
	logconfig "github.com/hyperledger/burrow/logging/logconfig"
)

type Config struct {
	Tendermint *tmConfig.TendermintConfig `toml:"tendermint"`
	RPC        *rpcConfig.RPCConfig       `toml:"rpc"`
	GRPC       *grpcConfig.GRPCConfig     `toml:"grpc"`
	Logging    *logconfig.LoggingConfig   `toml:"logging,omitempty"`
}

func DefaultConfig() *Config {
	return &Config{
		Tendermint: tmConfig.DefaultTendermintConfig(),
		RPC:        rpcConfig.DefaultRPCConfig(),
		GRPC:       grpcConfig.DefaultGRPCConfig(),
		Logging:    logconfig.DefaultNodeLoggingConfig(),
	}
}

func FromTOML(t string) (*Config, error) {
	conf := DefaultConfig()

	if _, err := toml.Decode(t, conf); err != nil {
		return nil, err
	}
	return conf, nil
}

func (conf *Config) ToTOML() ([]byte, error) {
	buf := new(bytes.Buffer)
	encoder := toml.NewEncoder(buf)
	err := encoder.Encode(conf)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func LoadFromFile(file string) (*Config, error) {
	dat, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	return FromTOML(string(dat))
}

func (conf *Config) SaveToFile(file string) error {
	toml, err := conf.ToTOML()
	if err != nil {
		return err
	}
	if err := common.WriteFile(file, toml); err != nil {
		return err
	}

	return nil
}
