package errors

import "errors"

var (
	ErrCryptoKeyGenerate         = errors.New("crypto key generation error")
	ErrCryptoKeyEncrypt          = errors.New("crypto key encryption error")
	ErrCryptoKeyDecrypt          = errors.New("crypto key decryption error")
	ErrUserPassHashing           = errors.New("user password hashing error")
	ErrConfigEnvParse            = errors.New("env vars parsing error")
	ErrMinioClientTransport      = errors.New("minio http transport error")
	ErrMinioClientCreate         = errors.New("minio http client create error")
	ErrObjectBucketAlreadyExists = errors.New("bucket already exists")
	ErrObjectBucketCreate        = errors.New("bucket creation error")
	ErrUserParse                 = errors.New("parse user string error")
	ErrAuthTokenNotFound         = errors.New("jwt token not found")
	ErrAuthTokenGenerate         = errors.New("jwt token generation error")
	ErrAuthTokenInvalid          = errors.New("jwt token invalid")
	ErrDBInit                    = errors.New("sql db init error")
	ErrDBConn                    = errors.New("sql db connection error")
	ErrDBClose                   = errors.New("sql db close error")
	ErrDBMigration               = errors.New("sql db migration error")
	ErrServerStart               = errors.New("grpc server start error")
	ErrServerShutdown            = errors.New("grpc server shutdown error")
	ErrServerTLS                 = errors.New("grpc server tls error")
	ErrTesting                   = errors.New("testing error")
)
