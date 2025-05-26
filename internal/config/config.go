package config

type Config struct {
	ServerAddr        string `env:"SERVER_ADDRESS"`
	ServerTLSKeyPath  string `env:"SERVER_TLS_KEY_PATH"`
	ServerTLSCertPath string `env:"SERVER_TLS_CERT_PATH"`
	S3Endpoint        string `env:"S3_ENDPOINT"`
	S3TLSCertPath     string `env:"S3_TLS_CERT_PATH"`
	S3AccessKey       string `env:"S3_ACCESS_KEY"`
	S3SecretKey       string `env:"S3_SECRET_KEY"`
	S3Token           string `env:"S3_TOKEN"`
}

func DefaultConfig() *Config {
	return &Config{
		ServerAddr:        `localhost:3200`,
		ServerTLSKeyPath:  `/etc/ssl/certs/gophkeeper/private.key`,
		ServerTLSCertPath: `/etc/ssl/certs/gophkeeper/public.crt`,
		S3Endpoint:        `localhost:9000`,
		S3TLSCertPath:     `/etc/ssl/certs/minio/public.crt`,
		S3AccessKey:       `minioadmin`,
		S3SecretKey:       `minioadmin`,
		S3Token:           ``,
	}
}

func LoadConfig() *Config {
	builder := newBuilder()
	cfg := builder.getConfig()

	return cfg
}
