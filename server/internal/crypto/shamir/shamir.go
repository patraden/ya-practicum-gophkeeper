package shamir

import "github.com/patraden/ya-practicum-gophkeeper/pkg/crypto/keys"

const (
	TotalShares     = 10
	ThresholdShares = 5
	ShareLength     = keys.REKLength + 1
)
