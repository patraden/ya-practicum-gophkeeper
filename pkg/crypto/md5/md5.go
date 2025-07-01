//nolint:gosec //reason: MD5 is used for non-security file integrity verification
package md5

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"

	e "github.com/patraden/ya-practicum-gophkeeper/pkg/errors"
)

// GetFileMD5 computes the MD5 hash of a file at the given path.
func GetFileMD5(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("[%w] md5 hash file", e.ErrRead)
	}
	defer file.Close()

	hasher := md5.New()

	if _, err := io.Copy(hasher, file); err != nil {
		return nil, fmt.Errorf("[%w] md5 hash file", e.ErrEncrypt)
	}

	return hasher.Sum(nil), nil
}
