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
