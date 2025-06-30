package stream

import (
	"fmt"
	"io"

	"github.com/minio/sio"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/crypto/keys"
	e "github.com/patraden/ya-practicum-gophkeeper/pkg/errors"
	"github.com/rs/zerolog"
)

// config returns a sio.Config configured with the given DEK and allowed cipher suites.
// It prefers AES-GCM and ChaCha20-Poly1305, both providing AEAD encryption.
func config(dek []byte) sio.Config {
	return sio.Config{
		CipherSuites: []byte{sio.AES_256_GCM, sio.CHACHA20_POLY1305},
		Key:          dek,
	}
}

// EncryptSecretStream creates a streaming encryption reader that reads plaintext from the input
// reader and returns an encrypted io.Reader. Encryption is done using AEAD (AES-256-GCM or ChaCha20-Poly1305).
// The function validates the provided DEK and returns a wrapped reader that encrypts on-the-fly.
func EncryptSecretStream(input io.Reader, dek []byte, log zerolog.Logger) (io.Reader, error) {
	if len(dek) != keys.DEKLength {
		log.Error().Msg("invalid DEK used for encryption")
		return nil, fmt.Errorf("[%w] invalid DEK length", e.ErrInvalidInput)
	}

	encryptedReader, err := sio.EncryptReader(input, config(dek))
	if err != nil {
		log.Error().Err(err).Msg("failed to create encryption stream")
		return nil, fmt.Errorf("[%w] secret encryption stream", e.ErrOpen)
	}

	log.Info().Msg("Secret ecnryption stream created successfully")

	return encryptedReader, nil
}

// DecryptSecretStream creates a streaming decryption reader that reads encrypted data from the input
// and returns a decrypted io.Reader. Decryption is performed using the same DEK used for encryption.
// The function validates the provided DEK and returns a wrapped reader that decrypts on-the-fly.
func DecryptSecretStream(input io.Reader, dek []byte, log zerolog.Logger) (io.Reader, error) {
	if len(dek) != keys.DEKLength {
		log.Error().Msg("invalid DEK used for decryption")
		return nil, fmt.Errorf("[%w] invalid DEK length", e.ErrInvalidInput)
	}

	decryptedReader, err := sio.DecryptReader(input, config(dek))
	if err != nil {
		log.Error().Err(err).Msg("failed to create decryption stream")
		return nil, fmt.Errorf("[%w] secret decryption stream", e.ErrOpen)
	}

	log.Info().Msg("Secret decryption stream created successfully")

	return decryptedReader, nil
}
