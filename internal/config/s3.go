package config

type S3Config struct {
	Endpoint  string `env:"OBJECT_STORAGE_ENDPOINT"`
	CertPath  string `env:"OBJECT_STORAGE_CERT_PATH"`
	AccessKey string `env:"OBJECT_STORAGE_ACCESS_KEY"`
	SecretKey string `env:"OBJECT_STORAGE_SECRET_KEY"`
	Token     string `env:"OBJECT_STORAGE_TOKEN"`
}

func DefaultObjectStore() *S3Config {
	return &S3Config{
		Endpoint:  `localhost:9000`,
		CertPath:  `.certs/minio-cert.crt`,
		AccessKey: `minioadmin`,
		SecretKey: `minioadmin`,
		Token:     ``,
	}
}

type Config struct {
	ObjectStorage S3Config
}

func DefaultConfig() *Config {
	return &Config{
		ObjectStorage: *DefaultObjectStore(),
	}
}

func LoadConfig() *Config {
	builder := newBuilder()
	cfg := builder.getConfig()

	return cfg
}
