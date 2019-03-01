package config

import (
	"fmt"

	"github.com/ethereum/go-ethereum/rpc"
)

type SputnikvmConfig struct {
	Web3Address string
}

func DefaultSputnikvmConfig() *SputnikvmConfig {
	return &SputnikvmConfig{
		Web3Address: "wss://mainnet.infura.io/ws",
	}
}

func (conf *SputnikvmConfig) Check() error {
	if conf.Web3Address == "" {
		return nil
	}

	// Make connection to the RPC client
	newRpcClient, err := rpc.Dial(conf.Web3Address)
	if err != nil {
		return fmt.Errorf("Failed to connect to the Ethereum network at '%v': %v", conf.Web3Address, err)
	}
	var output string
	err = newRpcClient.Call(&output, "web3_clientVersion")
	if err != nil {
		return fmt.Errorf("RPC call failed: %v", err)
	}

	fmt.Println("Gallactic successfully connected to Ethereum Network")

	return nil
}
