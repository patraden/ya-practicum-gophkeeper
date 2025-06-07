//nolint:err113 // reason: nested dynamic errors allowed in definitions.
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
	ErrUnavailable    = errors.New("service unavailable")
	ErrNotImplemented = errors.New("not implemented")
	ErrTimeout        = errors.New("timeout")
	ErrCanceled       = errors.New("operation canceled")
	ErrUnsupported    = errors.New("unsupported operation")
	ErrInternal       = InternalErr(errors.New("internal error"))

	// Access / Auth.
	ErrUnauthorized     = errors.New("unauthorized")
	ErrUnauthenticated  = errors.New("unauthenticated")
	ErrPermissionDenied = errors.New("permission denied")
	ErrForbidden        = errors.New("forbidden")
)

type InternalError struct {
	Err error
}

func InternalErr(err error) error {
	return &InternalError{Err: err}
}

func (e *InternalError) Error() string {
	if e.Err == nil {
		return "internal error"
	}

	return "internal error: " + e.Err.Error()
}

func (e *InternalError) Unwrap() error {
	return e.Err
}

func (e *InternalError) Is(target error) bool {
	_, ok := target.(*InternalError)
	return ok
}
