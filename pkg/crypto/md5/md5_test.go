package md5_test

import (
	"encoding/hex"
	"os"
	"testing"

	"github.com/patraden/ya-practicum-gophkeeper/pkg/crypto/md5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetFileMD5(t *testing.T) {
	t.Parallel()

	const (
		testContent    = "hello world"
		expectedMD5Hex = "5eb63bbbe01eeed093cb22bb8f5acdc3"
	)

	tmpFile, err := os.CreateTemp("", "hash_test_*.txt")
	require.NoError(t, err, "failed to create temp file")
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(testContent)
	require.NoError(t, err, "failed to write to temp file")

	require.NoError(t, tmpFile.Close(), "failed to close temp file")

	hashBytes, err := md5.GetFileMD5(tmpFile.Name())
	require.NoError(t, err, "GetFileMD5 should not return error")

	gotMD5Hex := hex.EncodeToString(hashBytes)
	assert.Equal(t, expectedMD5Hex, gotMD5Hex, "MD5 hash mismatch")
}
