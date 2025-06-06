package errors

import "errors"

var (
	// Validation and Input.
	ErrInvalidInput = errors.New("invalid input")
	ErrEmptyInput   = errors.New("empty input")
	ErrValidation   = errors.New("validation error")
	ErrGenerate     = errors.New("generation error")

	// Serialization.
	ErrParse     = errors.New("parse error")
	ErrMarshal   = errors.New("marshal error")
	ErrUnmarshal = errors.New("unmarshal error")
	ErrEncode    = errors.New("encode error")
	ErrDecode    = errors.New("decode error")

	// Security.
	ErrEncrypt = errors.New("encryption error")
	ErrDecrypt = errors.New("decryption error")

	// I/O.
	ErrOpen  = errors.New("open error")
	ErrRead  = errors.New("read error")
	ErrWrite = errors.New("write error")
	ErrSeek  = errors.New("seek error")
	ErrClose = errors.New("close error")

	// State and Existence.
	ErrNotFound = errors.New("not found")
	ErrExists   = errors.New("already exists")
	ErrConflict = errors.New("conflict")
	ErrCorrupt  = errors.New("corrupted data")
	ErrNotReady = errors.New("not ready")

	// Runtime / System.
	ErrInternal       = errors.New("internal error")
	ErrUnavailable    = errors.New("service unavailable")
	ErrNotImplemented = errors.New("not implemented")
	ErrRuntime        = errors.New("runtime error")
	ErrTimeout        = errors.New("timeout")
	ErrCanceled       = errors.New("operation canceled")
	ErrUnsupported    = errors.New("unsupported operation")

	// Access / Auth.
	ErrUnauthorized     = errors.New("unauthorized")
	ErrUnauthenticated  = errors.New("unauthenticated")
	ErrPermissionDenied = errors.New("permission denied")
	ErrForbidden        = errors.New("forbidden")
)
