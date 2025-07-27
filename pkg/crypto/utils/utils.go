package utils

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/binary"
	"fmt"
	"time"

	e "github.com/patraden/ya-practicum-gophkeeper/pkg/errors"
)

const tokenBytesShift = 16

func EqualHashes(a, b []byte) bool {
	return len(a) == len(b) && subtle.ConstantTimeCompare(a, b) == 1
}

// GenerateUploadToken creates a unique int64 token.
func GenerateUploadToken() (int64, error) {
	now := time.Now().UTC().Unix()

	var randBytes [2]byte
	if _, err := rand.Read(randBytes[:]); err != nil {
		return 0, fmt.Errorf("[%w] random bytes", e.ErrRead)
	}

	// Combine timestamp (high bits) and random (low bits)
	randomPart := binary.BigEndian.Uint16(randBytes[:])
	token := (now << tokenBytesShift) | int64(randomPart)

	return token, nil
}
