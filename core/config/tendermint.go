package config

type TendermintConfig struct {
	// Initial peers we connect to for peer exchange
	Seeds string `toml:"seeds"`
	// Peers to which we automatically connect
	PersistentPeers string `toml:"persistentPeers"`
	ListenAddress   string `toml:"listenAddress"`
	Moniker         string `toml:"moniker"`
	TendermintRoot  string `toml:"tendermintRoot"`
}

func DefaultTendermintConfig() TendermintConfig {
	return TendermintConfig{
		ListenAddress:  "tcp://0.0.0.0:46656",
		TendermintRoot: "./data",
	}
}
