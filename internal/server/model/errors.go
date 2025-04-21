package model

import "errors"

var (
	ErrConfigEnvParse            = errors.New("env vars parsing error")
	ErrMinioClientTransport      = errors.New("minio http transport error")
	ErrMinioClientCreate         = errors.New("minio http client create error")
	ErrObjectBucketAlreadyExists = errors.New("bucket already exists")
	ErrObjectBucketCreate        = errors.New("bucket creation error")
)
