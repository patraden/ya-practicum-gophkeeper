package stream_test

import (
	"bytes"
	"io"
	"testing"
	"time"

	"github.com/patraden/ya-practicum-gophkeeper/pkg/crypto/keys"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/crypto/stream"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/logger"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// bytePatternReader produces a repeating stream of bytes based on a fixed pattern.
// It's used to simulate input data without allocating large memory buffers.
type bytePatternReader struct {
	pattern []byte
	offset  int
}

func (r *bytePatternReader) Read(part []byte) (int, error) {
	for i := range part {
		part[i] = r.pattern[r.offset%len(r.pattern)]
		r.offset++
	}

	return len(part), nil
}

// countingSink is an io.Writer that discards written data but counts the total number of bytes written.
// It’s useful for benchmarking or validating stream size without keeping data in memory.
type countingSink struct {
	n *int64
}

func (s *countingSink) Write(p []byte) (int, error) {
	*s.n += int64(len(p))
	return len(p), nil
}

// TestEncryptDecryptStream verifies basic stream encryption/decryption of a small in-memory byte slice.
func TestEncryptDecryptStream(t *testing.T) {
	t.Parallel()

	log := logger.Stdout(zerolog.DebugLevel).GetZeroLog()
	dek, err := keys.DEK()
	require.NoError(t, err)

	originalData := []byte("this is test secret data, can be binary or text, maybe large")
	inputReader := bytes.NewReader(originalData)

	encryptedReader, err := stream.EncryptSecretStream(inputReader, dek, log)
	require.NoError(t, err)

	encryptedData, err := io.ReadAll(encryptedReader)
	require.NoError(t, err)

	assert.NotEqual(t, originalData, encryptedData, "encrypted data is different from original")

	decryptedReader, err := stream.DecryptSecretStream(bytes.NewReader(encryptedData), dek, log)
	require.NoError(t, err)

	decryptedData, err := io.ReadAll(decryptedReader)
	require.NoError(t, err)

	require.Equal(t, originalData, decryptedData, "encrypted/decrepted data matches original")
}

// TestEncryptDecryptStreamStress5GB performs a streaming encryption/decryption test with 5GB of synthetic data.
// It measures encryption and decryption time while ensuring memory safety and correctness.
func TestEncryptDecryptStreamStress5GB(t *testing.T) {
	t.Parallel()
	t.Skip("Skip by default to avoid long test runs. Remove to enable.")

	log := logger.Stdout(zerolog.DebugLevel).GetZeroLog()
	dek, err := keys.DEK()
	require.NoError(t, err)

	const totalSize int64 = 5 * 1024 * 1024 * 1024 // 5GB

	t.Logf("Generating %d bytes of data...", totalSize)

	// Reader that produces deterministic pseudorandom data
	dataReader := io.LimitReader(&bytePatternReader{pattern: []byte("gophkeeper-streaming-crypto-")}, totalSize)

	// Encrypt
	startEnc := time.Now()
	encReader, err := stream.EncryptSecretStream(dataReader, dek, log)
	require.NoError(t, err)

	// Read encrypted data into buffer
	var encryptedBuf bytes.Buffer
	nEnc, err := io.Copy(&encryptedBuf, encReader)
	require.NoError(t, err)
	t.Logf("Encrypted %d bytes in %v", nEnc, time.Since(startEnc))

	// Decrypt
	startDec := time.Now()
	decReader, err := stream.DecryptSecretStream(bytes.NewReader(encryptedBuf.Bytes()), dek, log)
	require.NoError(t, err)

	// Read decrypted data to /dev/null equivalent (just count bytes)
	var decryptedBytes int64
	sink := &countingSink{n: &decryptedBytes}
	_, err = io.Copy(sink, decReader)
	require.NoError(t, err)

	t.Logf("Decrypted %d bytes in %v", decryptedBytes, time.Since(startDec))
	require.Equal(t, totalSize, decryptedBytes, "decrypted data size should match original")
}
