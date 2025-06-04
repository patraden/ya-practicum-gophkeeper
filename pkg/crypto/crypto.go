package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha256"
	"io"

	"github.com/patraden/ya-practicum-gophkeeper/pkg/domain/user"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/errors"
)

const (
	kekLength = 32 // 256 bits
	dekLength = 32 // 256 bits
	kekIter   = 100_000
	nonceSize = 12 // nonceSize for AES-GCM (recommended size)
)

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
		return nil, errors.ErrCryptoKeyGenerate
	}

	kek, err := pbkdf2.Key(sha256.New, password, u.Salt, kekIter, kekLength)
	if err != nil {
		return nil, errors.ErrCryptoKeyGenerate
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
	dek := make([]byte, dekLength)
	if _, err := rand.Read(dek); err != nil {
		return nil, errors.ErrCryptoKeyGenerate
	}

	return dek, nil
}

// WrapDEK encrypts the given Data Encryption Key (DEK) using the provided Key Encryption Key (KEK).
// This function uses AES-GCM for authenticated encryption, ensuring both confidentiality and integrity.
// The resulting ciphertext includes a randomly generated nonce prepended to the encrypted DEK.
func WrapDEK(kek, dek []byte) ([]byte, error) {
	block, err := aes.NewCipher(kek)
	if err != nil {
		return nil, errors.ErrCryptoKeyEncrypt
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, errors.ErrCryptoKeyEncrypt
	}

	nonce := make([]byte, 0, nonceSize)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, errors.ErrCryptoKeyEncrypt
	}

	ciphertext := aesgcm.Seal(nil, nonce, dek, nil)

	return append(nonce, ciphertext...), nil
}

// UnwrapDEK decrypts a wrapped DEK using the given KEK.
// It expects the input to be nonce || ciphertext as returned by WrapDEK.
func UnwrapDEK(kek, wrapped []byte) ([]byte, error) {
	if len(wrapped) < nonceSize {
		return nil, errors.ErrCryptoKeyDecrypt
	}

	nonce := wrapped[:nonceSize]
	ciphertext := wrapped[nonceSize:]

	block, err := aes.NewCipher(kek)
	if err != nil {
		return nil, errors.ErrCryptoKeyDecrypt
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, errors.ErrCryptoKeyDecrypt
	}

	plaintext, err := aesgcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, errors.ErrCryptoKeyDecrypt
	}

	return plaintext, nil
}
