package app

import (
	"github.com/patraden/ya-practicum-gophkeeper/client/internal/config"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/logger"
)

func CreateSecret(cfg *config.Config, secretType, secretName, secretValue string, log logger.Logger) error {
	return nil
}

// func createSecret(
// 	cfg *config.Config,
// 	secretType,
// 	secretName string,
// 	secretValueReader io.Reader,
// 	log logger.Logger,
// ) error {
// 	return nil
// }
