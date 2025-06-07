package crypto

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
func (s *Splitter) Split(secret []byte, total, threshold int) ([][]byte, error) {
	shares, err := shamir.Split(secret, total, threshold)
	if err != nil {
		s.log.Error().Err(err).
			Int("total", total).
			Int("threshold", threshold).
			Msg("shamir's secret split")

		return nil, e.ErrInvalidInput
	}

	return shares, nil
}

// Combine reconstructs the original secret from a slice of shares.
func Combine(shares [][]byte) ([]byte, error) {
	secret, err := shamir.Combine(shares)
	if err != nil {
		return nil, e.ErrInvalidInput
	}

	return secret, nil
}
