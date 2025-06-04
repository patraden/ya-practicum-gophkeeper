package config

type Config struct {
	ServerAddr        string `env:"SERVER_ADDRESS"`
	ServerTLSCertPath string `env:"SERVER_TLS_CERT_PATH"`
	DatabaseDSN       string `env:"DATABASE_DSN"`
	InstallMode       bool
	DebugMode         bool
}

func DefaultConfig() *Config {
	return &Config{
		ServerAddr:        `localhost:3200`,
		ServerTLSCertPath: `/etc/ssl/certs/gophkeeper/backend/public.crt`,
		DatabaseDSN:       ``,
		InstallMode:       false,
		DebugMode:         false,
	}
}

func LoadConfig() *Config {
	builder := newBuilder()
	cfg := builder.getConfig()

	return cfg
}
