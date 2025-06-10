package shamir

import (
	"github.com/hashicorp/vault/shamir"
	e "github.com/patraden/ya-practicum-gophkeeper/pkg/errors"
	"github.com/rs/zerolog"
)

// Splitter splitts secrets into shares according to Shamir's Secret Sharing.
type Splitter struct {
	log *zerolog.Logger
}

// NewSplitter creates a new Splitter using the provided logger.
func NewSplitter(log *zerolog.Logger) *Splitter {
	return &Splitter{log: log}
}

// Split splits the secret into shares.
func (s *Splitter) Split(secret []byte) ([][]byte, error) {
	shares, err := shamir.Split(secret, TotalShares, ThresholdShares)
	if err != nil {
		s.log.Error().Err(err).
			Int("total", TotalShares).
			Int("threshold", ThresholdShares).
			Msg("shamir's secret split")

		return nil, e.ErrInvalidInput
	}

	return shares, nil
}
