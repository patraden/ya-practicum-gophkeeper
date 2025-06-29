package s3

import (
	pb "github.com/patraden/ya-practicum-gophkeeper/pkg/proto/gophkeeper/v1"
)

// TemporaryCredentials represents temporary security credentials
// issued by an STS-compatible identity provider (e.g., AWS STS or MinIO).
type TemporaryCredentials struct {
	AccessKeyID     string `xml:"AccessKeyId"`
	SecretAccessKey string `xml:"SecretAccessKey"`
	SessionToken    string `xml:"SessionToken"`
	Expiration      string `xml:"Expiration"`
}

func (creds TemporaryCredentials) ToProto() *pb.TemporaryCredentials {
	return &pb.TemporaryCredentials{
		AccessKeyId:     creds.AccessKeyID,
		SecretAccessKey: creds.SecretAccessKey,
		SessionToken:    creds.SessionToken,
		Expiration:      creds.Expiration,
	}
}
