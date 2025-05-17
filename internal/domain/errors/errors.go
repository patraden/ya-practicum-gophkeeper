package errors

import "errors"

var (
	ErrConfigEnvParse            = errors.New("env vars parsing error")
	ErrMinioClientTransport      = errors.New("minio http transport error")
	ErrMinioClientCreate         = errors.New("minio http client create error")
	ErrObjectBucketAlreadyExists = errors.New("bucket already exists")
	ErrObjectBucketCreate        = errors.New("bucket creation error")
	ErrUserParse                 = errors.New("parse user string error")
	ErrJWTTokenNotFound          = errors.New("jwt token not found")
	ErrJWTTokenGenerate          = errors.New("jwt token generation error")
	ErrJWTTokenInvalid           = errors.New("jwt token invalid")
	ErrTesting                   = errors.New("testing error")
)
