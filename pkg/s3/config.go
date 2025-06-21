package s3

type ClientConfig struct {
	S3Endpoint    string `env:"S3_ENDPOINT"`
	S3TLSCertPath string `env:"S3_TLS_CERT_PATH"`
	S3AccessKey   string `env:"S3_ACCESS_KEY"`
	S3SecretKey   string `env:"S3_SECRET_KEY"`
	S3AccountID   string `env:"S3_ACCOUNT_ID"`
	S3Region      string `env:"S3_REGION"`
	S3Token       string `env:"S3_TOKEN"`
}

func DefaultClientConfig() *ClientConfig {
	return &ClientConfig{
		S3Endpoint:    `localhost:9000`,
		S3TLSCertPath: `/etc/ssl/certs/gophkeeper/minio/public.crt`,
		S3AccessKey:   `gophkeeper`,
		S3SecretKey:   `gophkeeper`,
		S3Token:       ``,
		S3AccountID:   `gophkeeper`,
		S3Region:      `eu-central-1`,
	}
}
