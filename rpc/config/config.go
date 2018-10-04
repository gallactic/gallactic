package config

type RPCConfig struct {
	Enabled bool
	Server  *ServerConfig
}

type GRPCConfig struct {
	Enabled    bool
	GRPCserver *ServerConfig
}

func DefaultGRPCConfig() *GRPCConfig {
	return &GRPCConfig{
		Enabled:    true,
		GRPCserver: DefaultGRPCServerConfig(),
	}
}

func DefaultRPCConfig() *RPCConfig {
	return &RPCConfig{
		Enabled: true,
		Server:  DefaultServerConfig(),
	}
}

type (
	ServerConfig struct {
		Bind Bind `toml:"bind"`
		TLS  TLS  `toml:"TLS"`
		CORS CORS `toml:"CORS"`
		HTTP HTTP `toml:"HTTP"`
	}

	Bind struct {
		Address string `toml:"address"`
		Port    uint16 `toml:"port"`
	}

	TLS struct {
		TLS      bool   `toml:"tls"`
		CertPath string `toml:"cert_path"`
		KeyPath  string `toml:"key_path"`
	}

	// Options stores configurations
	CORS struct {
		Enable           bool     `toml:"enable"`
		AllowOrigins     []string `toml:"allow_origins"`
		AllowCredentials bool     `toml:"allow_credentials"`
		AllowMethods     []string `toml:"allow_methods"`
		AllowHeaders     []string `toml:"allow_headers"`
		ExposeHeaders    []string `toml:"expose_headers"`
		MaxAge           uint64   `toml:"max_age"`
	}

	HTTP struct {
		JsonRpcEndpoint string `toml:"json_rpc_endpoint"`
	}
)

func DefaultServerConfig() *ServerConfig {
	cp := ""
	kp := ""
	return &ServerConfig{
		Bind: Bind{
			Address: "localhost",
			Port:    1337,
		},
		TLS: TLS{TLS: false,
			CertPath: cp,
			KeyPath:  kp,
		},
		CORS: CORS{},
		HTTP: HTTP{
			JsonRpcEndpoint: "/rpc",
		},
	}
}

func DefaultGRPCServerConfig() *ServerConfig {
	return &ServerConfig{
		Bind: Bind{
			Address: "localhost",
			Port:    10903,
		},
	}
}
