package s3

// TemporaryCredentials represents temporary security credentials
// issued by an STS-compatible identity provider (e.g., AWS STS or MinIO).
type TemporaryCredentials struct {
	AccessKeyID     string `xml:"AccessKeyId"`
	SecretAccessKey string `xml:"SecretAccessKey"`
	SessionToken    string `xml:"SessionToken"`
	Expiration      string `xml:"Expiration"`
}
