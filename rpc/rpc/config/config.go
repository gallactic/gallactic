package config

type Bind struct {
	Address string `toml:"address"`
	Port    uint16 `toml:"port"`
}

type TLS struct {
	TLS      bool   `toml:"tls"`
	CertPath string `toml:"cert_path"`
	KeyPath  string `toml:"key_path"`
}

type CORS struct {
	Enabled          bool     `toml:"enable"`
	AllowOrigins     []string `toml:"allow_origins"`
	AllowCredentials bool     `toml:"allow_credentials"`
	AllowMethods     []string `toml:"allow_methods"`
	AllowHeaders     []string `toml:"allow_headers"`
	ExposeHeaders    []string `toml:"expose_headers"`
	MaxAge           uint64   `toml:"max_age"`
}

type HTTP struct {
	JsonRpcEndpoint string `toml:"json_rpc_endpoint"`
}

type ServerConfig struct {
	Bind Bind `toml:"bind"`
	TLS  TLS  `toml:"TLS"`
	CORS CORS `toml:"CORS"`
	HTTP HTTP `toml:"HTTP"`
}

type RPCConfig struct {
	Enabled bool
	Server  *ServerConfig
}

func DefaultRPCConfig() *RPCConfig {
	return &RPCConfig{
		Enabled: true,
		Server: &ServerConfig{
			Bind: Bind{
				Address: "0.0.0.0",
				Port:    1337,
			},
			TLS: TLS{TLS: false,
				CertPath: "",
				KeyPath:  "",
			},
			CORS: CORS{},
			HTTP: HTTP{
				JsonRpcEndpoint: "/rpc",
			},
		},
	}
}
