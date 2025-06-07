package keys

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"

	"github.com/patraden/ya-practicum-gophkeeper/pkg/domain/user"
	e "github.com/patraden/ya-practicum-gophkeeper/pkg/errors"
)

// Length constants for all key types in bytes.
const (
	REKLength      = 32 // Root Encryption Key (256-bit)
	KEKLength      = 32 // Key Encryption Key (256-bit)
	DEKLength      = 32 // Data Encryption Key (256-bit)
	kekIter        = 100_000
	nonceSize      = 12 // Recommended nonce size for AES-GCM
	EncryptionAlgo = "AES-GCM"
)

// REK generates a secure random Root Encryption Key (REK).
// This key is intended to be split using Shamir's Secret Sharing algorithm
// and securely stored across multiple trusted parties.
//
// The REK serves as the root of trust for protecting secrets stored
// in GophKeeper and should never be persisted as-is.
func REK() ([]byte, error) {
	rek := make([]byte, REKLength)
	if _, err := rand.Read(rek); err != nil {
		return nil, e.InternalErr(err)
	}

	return rek, nil
}

// KEK derives a Key Encryption Key (KEK) from the user's password and stored salt.
// In the context of GophKeeper, the KEK serves as the user's master key —
// a symmetric cryptographic key deterministically derived from the user's password.
// It is never stored and is re-derived at runtime when needed.
//
// This key is used exclusively on the client side to:
//   - Decrypt Data Encryption Keys (DEKs), which are used to encrypt/decrypt actual user data.
//   - Secure data at rest before uploading to the server.
//   - Ensure end-to-end encryption, where the server stores encrypted data but cannot decrypt it.
func KEK(u *user.User, password string) ([]byte, error) {
	if !u.CheckPassword(password) {
		return nil, e.ErrInvalidInput
	}

	kek, err := pbkdf2.Key(sha256.New, password, u.Salt, kekIter, KEKLength)
	if err != nil {
		return nil, e.ErrGenerate
	}

	return kek, nil
}

// DEK generates a new random 256-bit (32-byte) Data Encryption Key (DEK).
// In the context of GophKeeper, the DEK is used by the client to encrypt user secrets
// before they are uploaded to the server.
//
// Unlike the KEK (which is derived from the user's password and used to wrap/unwrap DEKs),
// the DEK is randomly generated and unique per secret.
//
// This approach ensures forward secrecy: if one DEK is ever compromised,
// it doesn't affect the security of other encrypted secrets.
func DEK() ([]byte, error) {
	dek := make([]byte, DEKLength)
	if _, err := rand.Read(dek); err != nil {
		return nil, e.ErrGenerate
	}

	return dek, nil
}

// HashREK calculates a SHA-256 hash of the given Root Encryption Key (REK).
// This hash can be stored securely and used later to verify the correctness
// of the reconstructed REK without persisting the original key.
func HashREK(rek []byte) []byte {
	hash := sha256.Sum256(rek)
	return hash[:]
}

// WrapKEK encrypts a KEK with the REK using the same AES-GCM mechanism.
func WrapKEK(rek, kek []byte) ([]byte, error) {
	return WrapDEK(rek, kek)
}

// UnwrapKEK decrypts a wrapped KEK using the REK.
func UnwrapKEK(rek, wrapped []byte) ([]byte, error) {
	return UnwrapDEK(rek, wrapped)
}

// WrapDEK encrypts the given Data Encryption Key (DEK) using the provided Key Encryption Key (KEK).
// This function uses AES-GCM for authenticated encryption, ensuring both confidentiality and integrity.
// The resulting ciphertext includes a randomly generated nonce prepended to the encrypted DEK.
func WrapDEK(kek, dek []byte) ([]byte, error) {
	if len(kek) != KEKLength || len(dek) != DEKLength {
		return nil, e.ErrInvalidInput
	}

	block, err := aes.NewCipher(kek)
	if err != nil {
		return nil, fmt.Errorf("encrypt dek(cipher): %w", e.ErrEncrypt)
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("encrypt dek(gcm): %w", e.ErrEncrypt)
	}

	nonce := make([]byte, nonceSize)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("encrypt dek(nonce): %w", e.ErrEncrypt)
	}

	ciphertext := aesgcm.Seal(nil, nonce, dek, nil)
	result := make([]byte, nonceSize+len(ciphertext))

	copy(result, nonce)
	copy(result[nonceSize:], ciphertext)

	return result, nil
}

// UnwrapDEK decrypts a wrapped DEK using the given KEK.
// It expects the input to be nonce || ciphertext as returned by WrapDEK.
func UnwrapDEK(kek, wrapped []byte) ([]byte, error) {
	if len(kek) != KEKLength {
		return nil, e.ErrInvalidInput
	}

	if len(wrapped) < nonceSize {
		return nil, e.ErrInvalidInput
	}

	nonce := wrapped[:nonceSize]
	ciphertext := wrapped[nonceSize:]

	block, err := aes.NewCipher(kek)
	if err != nil {
		return nil, fmt.Errorf("decrypt dek(cipher): %w", e.ErrDecrypt)
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("decrypt dek(gcm): %w", e.ErrDecrypt)
	}

	plaintext, err := aesgcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decrypt dek(gcm open): %w", e.ErrDecrypt)
	}

	return plaintext, nil
}
