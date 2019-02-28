package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/gallactic/gallactic/common"
	sputnikvmConfig "github.com/gallactic/gallactic/core/evm/sputnikvm/config"
	grpcConfig "github.com/gallactic/gallactic/www/grpc/config"
	rpcConfig "github.com/gallactic/gallactic/www/rpc/config"
	tmConfig "github.com/tendermint/tendermint/config"
)

type Config struct {
	Tendermint *tmConfig.Config                 `toml:"Tendermint"`
	RPC        *rpcConfig.RPCConfig             `toml:"RPC"`
	GRPC       *grpcConfig.GRPCConfig           `toml:"GRPC"`
	Logging    *Logging                         `toml:"Logging,omitempty"`
	SputnikVM  *sputnikvmConfig.SputnikvmConfig `toml:"SputnikVM"`
}

func DefaultConfig() *Config {
	tmDef := tmConfig.DefaultConfig()
	tmDef.SetRoot("./data")
	tmDef.P2P.ListenAddress = "tcp://0.0.0.0:46656"
	tmDef.RPC.ListenAddress = "tcp://localhost:46657"
	tmDef.Consensus.TimeoutCommit = 5 * tmDef.Consensus.TimeoutCommit
	tmDef.Consensus.CreateEmptyBlocks = false
	tmDef.ProxyApp = "gallactic"
	tmDef.PrivValidatorKey = ""
	tmDef.Genesis = ""

	return &Config{
		Tendermint: tmDef,
		RPC:        rpcConfig.DefaultRPCConfig(),
		GRPC:       grpcConfig.DefaultGRPCConfig(),
		SputnikVM:  sputnikvmConfig.DefaultSputnikvmConfig(),
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

func FromJSON(t string) (*Config, error) {
	conf := DefaultConfig()
	if err := json.Unmarshal([]byte(t), conf); err != nil {
		return nil, err
	}
	return conf, nil
}

func (conf *Config) ToJSON() ([]byte, error) {
	return json.MarshalIndent(conf, "", "  ")
}

func LoadFromFile(file string) (*Config, error) {
	dat, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	if strings.HasSuffix(file, "toml") {
		return FromTOML(string(dat))
	} else if strings.HasSuffix(file, "json") {
		return FromJSON(string(dat))
	}

	return nil, errors.New("Invalid suffix for the config file")
}

func (conf *Config) SaveToFile(file string) error {
	var dat []byte
	if strings.HasSuffix(file, "toml") {
		dat, _ = conf.ToTOML()
	} else if strings.HasSuffix(file, "json") {
		dat, _ = conf.ToJSON()
	} else {
		return errors.New("Invalid suffix for the config file")
	}
	if err := common.WriteFile(file, dat); err != nil {
		return err
	}

	return nil
}

// Verify web3 connection - to use it in interChainTrx precompiled contract to connect
func (conf *Config) Check() error {
	return conf.SputnikVM.Check()
}
