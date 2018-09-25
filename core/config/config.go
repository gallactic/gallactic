package config

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	tmConfig "github.com/gallactic/gallactic/core/consensus/tendermint/config"
	rpcConfig "github.com/gallactic/gallactic/rpc/config"
	logconfig "github.com/hyperledger/burrow/logging/logconfig"
)

type Config struct {
	Tendermint *tmConfig.TendermintConfig `toml:"tendermint"`
	RPC        *rpcConfig.RPCConfig       `toml:"rpc"`
	Logging    *logconfig.LoggingConfig   `toml:"logging,omitempty"`
}

func defaultConfig() *Config {
	return &Config{
		Tendermint: tmConfig.DefaultTendermintConfig(),
		RPC:        rpcConfig.DefaultRPCConfig(),
		Logging:    logconfig.DefaultNodeLoggingConfig(),
	}
}

func LoadFromFile(file string) (*Config, error) {
	fmt.Println("file", file)
	dat, err := ioutil.ReadFile(file)
	fmt.Println("dat", dat)
	if err != nil {
		return nil, err
	}
	return FromTOML(string(dat))
}


func SaveConfigFile(workingDir string) string {

	/*check for working path */
	if workingDir == "" {
		workingDir = "/tmp/chain/"
	}
	configpath := workingDir + "config.toml"
	var config = defaultConfig()
	conf := config.ToTOML()
	if err := os.MkdirAll(filepath.Dir(configpath), 0700); err != nil {
		log.Fatalf("Could not create directory %s", filepath.Dir(configpath))
	}
	if err := ioutil.WriteFile(configpath, []byte(conf), 0600); err != nil {
		log.Fatalf("Failed to write config file to %s: %v", configpath, err)
	}
	msg := " The file has created at " 
	return msg
}

func FromTOML(t string) (*Config, error) {
	conf := defaultConfig()

	if _, err := toml.Decode(t, conf); err != nil {
		return nil, err
	}
	return conf, nil
}

func (conf *Config) ToTOML() string {
	buf := new(bytes.Buffer)
	encoder := toml.NewEncoder(buf)
	err := encoder.Encode(conf)
	if err != nil {
		return fmt.Sprintf("Could not serialize config: %v", err)
	}

	return buf.String()
}
