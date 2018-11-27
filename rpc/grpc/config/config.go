package config

import (
	"fmt"
)

const localhost = "127.0.0.1"

type GRPCConfig struct {
	Enabled       bool
	ListenAddress string
}

func DefaultGRPCConfig() *GRPCConfig {
	return &GRPCConfig{
		Enabled:       true,
		ListenAddress: fmt.Sprintf("%s:50051", localhost),
	}
}
