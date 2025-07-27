//go:generate easyjson -all config.go

package config

import "time"

const (
	ConfigFileName    = "gophkeeper.json"
	DefaultReqTimeout = 180 * time.Second
	DefaultServerPort = 3200
)

// Config holds CLI settings and can be serialized to/from JSON.
type Config struct {
	InstallDir        string `env:"INSTALL_DIR"             json:"install_dir"`
	ServerHost        string `env:"SERVER_HOST"             json:"server_host"`
	ServerPort        int    `env:"SERVER_PORT"             json:"server_port"`
	ServerTLSCertPath string `env:"SERVER_TLS_CERT_PATH"    json:"server_tls_cert_path"`
	DatabaseFileName  string `env:"DATABASE_FILE_NAME"      json:"database_dsn"`
	S3Endpoint        string `env:"S3_ENDPOINT"             json:"s3_endpoint"`
	S3AccountID       string `env:"S3_ACCOUNT_ID"           json:"s3_account_id"`
	S3Region          string `env:"S3_REGION"               json:"s3_region"`
	Username          string `env:"GOPHKEEPER_USERNAME"     json:"-"`
	Password          string `env:"GOPHKEEPER_USERPASSWORD" json:"-"`
	DebugMode         bool   `env:"DEBUG"                   json:"debug"`
	InstallMode       bool   `json:"-"`
	RequestsTimeout   time.Duration
}

func DefaultConfig() *Config {
	return &Config{
		InstallDir:        `.gophkeeper`,
		ServerHost:        `localhost`,
		ServerPort:        DefaultServerPort,
		ServerTLSCertPath: `./deployments/.certs/ca.cert`,
		DatabaseFileName:  `gophkeeper.db`,
		S3Endpoint:        `localhost:9000`,
		S3AccountID:       `gophkeeper`,
		S3Region:          `eu-central-1`,
		Username:          ``,
		Password:          ``,
		RequestsTimeout:   DefaultReqTimeout,
		InstallMode:       false,
		DebugMode:         false,
	}
}

func LoadConfig(dcfg *Config) *Config {
	builder := newBuilder(dcfg)
	cfg := builder.getConfig()

	return cfg
}
