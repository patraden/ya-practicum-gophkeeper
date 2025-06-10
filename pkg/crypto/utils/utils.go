package utils

import "crypto/subtle"

func EqualHashes(a, b []byte) bool {
	return len(a) == len(b) && subtle.ConstantTimeCompare(a, b) == 1
}
