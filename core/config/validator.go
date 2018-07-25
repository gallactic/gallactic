package config

import "github.com/gallactic/gallactic/crypto"

type ValidatorConfig struct {
	Address crypto.Address `toml:"address"`
}

func DefaultValidatorConfig() *ValidatorConfig {
	return &ValidatorConfig{}
}
