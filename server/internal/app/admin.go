// Package app contains application-level use cases and business logic,
// including administrative operations such as unsealing the server's REK (Root Encryption Key).
package app

import (
	"context"
	"errors"

	"github.com/patraden/ya-practicum-gophkeeper/pkg/crypto/keys"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/crypto/utils"
	e "github.com/patraden/ya-practicum-gophkeeper/pkg/errors"
	pb "github.com/patraden/ya-practicum-gophkeeper/pkg/proto/gophkeeper/v1"
	"github.com/patraden/ya-practicum-gophkeeper/server/internal/crypto/keystore"
	"github.com/patraden/ya-practicum-gophkeeper/server/internal/crypto/shamir"
	"github.com/patraden/ya-practicum-gophkeeper/server/internal/repository"
	"github.com/rs/zerolog"
)

const (
	StatusUnspecified = pb.SealStatus_SEAL_STATUS_UNSPECIFIED
	StatusSealed      = pb.SealStatus_SEAL_STATUS_SEALED
	StatusUnsealed    = pb.SealStatus_SEAL_STATUS_UNSEALED
)

// AdminUseCase defines administrative operations available for server control.
type AdminUseCase interface {
	// Unseal processes a Shamir share and attempts to unseal the server.
	// Returns the current seal status and a user-friendly message.
	Unseal(ctx context.Context, share []byte) (pb.SealStatus, string)
}

// AdminUC implements AdminUseCase. It orchestrates the REK unsealing logic
// using a Shamir share collector and secure keystore, validated against a stored hash.
type AdminUC struct {
	AdminUseCase
	collector *shamir.Collector        // Used to collect and reconstruct the REK
	kstore    keystore.Keystore        // Secure memory-backed store for the REK
	repo      repository.REKRepository // Interface to access REK hash stored in the database
	log       zerolog.Logger
}

// NewAdminUC creates a new instance of AdminUC.
func NewAdminUC(
	collector *shamir.Collector,
	kstore keystore.Keystore,
	repo repository.REKRepository,
	log zerolog.Logger,
) *AdminUC {
	return &AdminUC{
		collector: collector,
		kstore:    kstore,
		repo:      repo,
		log:       log,
	}
}

// Unseal processes a single base64-decoded Shamir share.
// If enough valid shares are collected, the REK is reconstructed, verified via hash,
// and stored securely in memory. The function returns the current seal status
// and a human-readable message.
func (uc *AdminUC) Unseal(ctx context.Context, share []byte) (pb.SealStatus, string) {
	if uc.kstore.IsLoaded() {
		return StatusUnsealed, "Unsealed previously"
	}

	if err := uc.collector.Collect(share); err != nil {
		if errors.Is(err, e.ErrConflict) {
			uc.log.Info().
				Int("size", uc.collector.Size()).
				Msg("Collector already full")
		} else {
			uc.log.Error().Err(err).
				Int("size", uc.collector.Size()).
				Msg("Failed to collect share")

			return StatusSealed, "Failed to collect provided key piece"
		}
	}

	uc.log.Info().Msg("Reconstructing REK from shares")

	rek, err := uc.collector.Reconstruct()
	if errors.Is(err, e.ErrNotReady) {
		return StatusSealed, uc.collector.StatusMessage()
	}

	if err != nil {
		uc.log.Error().Err(err).Msg("Failed to reconstruct REK")
		uc.collector.Reset()

		return StatusSealed, "Bad root key pieces collected. All key pieces wiped."
	}

	expectedHash, err := uc.repo.GetHash(ctx)
	if err != nil {
		uc.log.Error().Err(err).Msg("Failed to retrieve REK hash for validation")

		return StatusSealed, "Internal error during root key validation: " + err.Error()
	}

	if !utils.EqualHashes(keys.HashREK(rek), expectedHash) {
		uc.log.Error().Msg("REK validation failed")
		uc.collector.Reset()

		return StatusSealed, "Bad root key provided. All key pieces wiped."
	}

	if err := uc.kstore.Load(rek); err != nil {
		uc.log.Error().Err(err).Msg("Failed to load REK into keystore")

		return StatusSealed, "Internal error during root key store: " + err.Error()
	}

	uc.log.Info().Msg("REK successfully reconstructed and stored")

	return StatusUnsealed, "Unsealed now"
}
