package config

type Config struct {
	ServerAddr           string `env:"SERVER_ADDRESS"`
	ServerTLSKeyPath     string `env:"SERVER_TLS_KEY_PATH"`
	ServerTLSCertPath    string `env:"SERVER_TLS_CERT_PATH"`
	DatabaseDSN          string `env:"DATABASE_DSN"`
	S3Endpoint           string `env:"S3_ENDPOINT"`
	S3TLSCertPath        string `env:"S3_TLS_CERT_PATH"`
	S3AccessKey          string `env:"S3_ACCESS_KEY"`
	S3SecretKey          string `env:"S3_SECRET_KEY"`
	S3AccountID          string `env:"S3_ACCOUNT_ID"`
	S3Region             string `env:"S3_REGION"`
	S3RedisRegion        string `env:"S3_REDIS_REGION"`
	S3Token              string `env:"S3_TOKEN"`
	IdentityTLSCertPath  string `env:"IDENTITY_TLS_CERT_PATH"`
	IdentityEndpoint     string `env:"IDENTITY_ENDPOINT"`
	IdentityClientID     string `env:"IDENTITY_OPENID_CLIENT_ID"`
	IdentityClientSecret string `env:"IDENTITY_OPENID_CLIENT_SECRET"`
	IdentityRealm        string `env:"IDENTITY_OPENID_REALM"`
	JWTSecret            string `env:"JWT_SECRET"`
	REKSharesPath        string `env:"REK_SHARES_PATH"`
	InstallMode          bool
	DebugMode            bool
}

func DefaultConfig() *Config {
	return &Config{
		ServerAddr:           `localhost:3200`,
		ServerTLSKeyPath:     `/etc/ssl/certs/gophkeeper/backend/private.key`,
		ServerTLSCertPath:    `/etc/ssl/certs/gophkeeper/backend/public.crt`,
		DatabaseDSN:          ``,
		S3Endpoint:           `localhost:9000`,
		S3TLSCertPath:        `/etc/ssl/certs/gophkeeper/minio/public.crt`,
		S3AccessKey:          `gophkeeper`,
		S3SecretKey:          `gophkeeper`,
		S3Token:              ``,
		S3AccountID:          `gophkeeper`,
		S3Region:             `eu-central-1`,
		S3RedisRegion:        `eu-central-1`,
		IdentityTLSCertPath:  ``,
		IdentityEndpoint:     `http://localhost:8080`,
		IdentityClientID:     `minio`,
		IdentityClientSecret: `FRwPgFU1i6reAgN6lH2iM4qgQZeAMjSv`,
		IdentityRealm:        `gophkeeper`,
		JWTSecret:            `d1a58c288a0226998149277b14993f6c73cf44ff9df3de548df4df25a13b251a`,
		REKSharesPath:        `shares.json`,
		InstallMode:          false,
		DebugMode:            false,
	}
}

func LoadConfig() *Config {
	builder := newBuilder()
	cfg := builder.getConfig()

	return cfg
}
