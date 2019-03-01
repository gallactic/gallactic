package config

type Bind struct {
	Address string
	Port    uint16
}

type TLS struct {
	TLS      bool
	CertPath string
	KeyPath  string
}

type CORS struct {
	Enabled          bool
	AllowOrigins     []string
	AllowCredentials bool
	AllowMethods     []string
	AllowHeaders     []string
	ExposeHeaders    []string
	MaxAge           uint64
}

type HTTP struct {
	JsonRpcEndpoint string
}

type ServerConfig struct {
	Bind Bind
	TLS  TLS
	CORS CORS
	HTTP HTTP
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
