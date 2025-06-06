package crypto

import (
	"github.com/hashicorp/vault/shamir"
	e "github.com/patraden/ya-practicum-gophkeeper/pkg/errors"
	"github.com/rs/zerolog"
)

type Splitter interface {
	SplitAndDistribute(secret []byte, total, threshold int) ([][]byte, error)
}

type StdoutSplitter struct {
	log *zerolog.Logger
}

func NewStdoutSplitter(log *zerolog.Logger) *StdoutSplitter {
	return &StdoutSplitter{log: log}
}

// SplitAndDistribute splits the secret and logs the shares.
// Returns shares for testing or further processing.
func (s *StdoutSplitter) SplitAndDistribute(secret []byte, total, threshold int) ([][]byte, error) {
	shares, err := shamir.Split(secret, total, threshold)
	if err != nil {
		s.log.Error().Err(err).
			Msg("shamir's secret split")

		return nil, e.ErrInternal
	}

	for i, share := range shares {
		s.log.Info().
			Msgf("Share %d: %x", i+1, share)
	}

	return shares, nil
}

// Combine reconstructs the secret from a set of shares.
func Combine(shares [][]byte) ([]byte, error) {
	secret, err := shamir.Combine(shares)
	if err != nil {
		return nil, e.ErrInvalidInput
	}

	return secret, nil
}
