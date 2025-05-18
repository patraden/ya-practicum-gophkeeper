package errors

import "errors"

var (
	ErrConfigEnvParse            = errors.New("env vars parsing error")
	ErrMinioClientTransport      = errors.New("minio http transport error")
	ErrMinioClientCreate         = errors.New("minio http client create error")
	ErrObjectBucketAlreadyExists = errors.New("bucket already exists")
	ErrObjectBucketCreate        = errors.New("bucket creation error")
	ErrUserParse                 = errors.New("parse user string error")
	ErrAuthTokenNotFound         = errors.New("jwt token not found")
	ErrAuthTokenGenerate         = errors.New("jwt token generation error")
	ErrAuthTokenInvalid          = errors.New("jwt token invalid")
	ErrPGInit                    = errors.New("pg database init error")
	ErrPGConn                    = errors.New("pg database connection error")
	ErrTesting                   = errors.New("testing error")
)
