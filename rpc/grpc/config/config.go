package config

import (
	"fmt"
)

const localhost = "0.0.0.0"

type GRPCConfig struct {
	Enabled       bool
	ListenAddress string
	HTTPAddress   string
}

func DefaultGRPCConfig() *GRPCConfig {
	return &GRPCConfig{
		Enabled:       true,
		ListenAddress: fmt.Sprintf("%s:50051", localhost),
		HTTPAddress: fmt.Sprintf("%s:50052", localhost),
	}
}
