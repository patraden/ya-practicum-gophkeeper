package auth

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
)

// GenerateVerifier returns an HMAC-based verifier using the user's password and a salt.
// This verifier is a byte slice and can be sent to the client for authentication challenge/response.
func GenerateVerifier(password string, salt []byte) []byte {
	mac := hmac.New(sha256.New, []byte(password))
	mac.Write(salt)

	return mac.Sum(nil)
}

// VerifyVerifier checks whether the given password and salt generate the expected verifier.
func VerifyVerifier(password string, salt []byte, expected []byte) bool {
	return bytes.Equal(GenerateVerifier(password, salt), expected)
}
