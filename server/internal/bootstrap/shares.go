package bootstrap

import (
	"os"

	"github.com/patraden/ya-practicum-gophkeeper/pkg/dto"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/errors"
	"github.com/rs/zerolog"
)

func WriteSharesFile(shares [][]byte, path string, log *zerolog.Logger) error {
	out := dto.ShamirShares{Shares: shares}

	file, err := os.Create(path)
	if err != nil {
		log.Error().Err(err).
			Str("file", path).
			Msg("failed to open file for shares")

		return errors.ErrOpen
	}
	defer file.Close()

	data, err := out.MarshalJSON()
	if err != nil {
		log.Error().Err(err).
			Str("file", path).
			Msg("failed to marshal shares")

		return errors.ErrMarshal
	}

	if _, err := file.Write(data); err != nil {
		log.Error().Err(err).
			Str("file", path).
			Msg("failed to writes shares to file")

		return errors.ErrWrite
	}

	log.Info().
		Str("file", path).
		Msg("shares written")

	return nil
}
