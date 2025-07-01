package cmd

import (
	"fmt"

	"github.com/patraden/ya-practicum-gophkeeper/client/internal/app"
	"github.com/patraden/ya-practicum-gophkeeper/client/internal/config"
	e "github.com/patraden/ya-practicum-gophkeeper/pkg/errors"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/logger"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

func NewCreateCmd(dcfg *config.Config) *cobra.Command {
	log := logger.StdoutConsole(zerolog.DebugLevel)

	var (
		secretName  string
		secretType  string
		secretValue string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Creates a user secret in GophKeeper",
		RunE: func(_ *cobra.Command, _ []string) error {
			if secretName == "" {
				return fmt.Errorf("[%w] --secret flag is required", e.ErrInvalidInput)
			}
			switch secretType {
			case "binary", "card", "credentials":
			default:
				return fmt.Errorf("[%w] --type must be one of: binary, card, credentials", e.ErrInvalidInput)
			}
			if secretValue == "" {
				return fmt.Errorf("[%w] --value must be provided for secret", e.ErrInvalidInput)
			}

			cfg := config.LoadConfig(dcfg)
			return app.CreateSecret(cfg, secretType, secretName, secretValue, log)
		},
		SilenceUsage: true,
	}

	cmd.Flags().StringVarP(&dcfg.Username, "username", "u", dcfg.Username, "Username")
	cmd.Flags().StringVarP(&dcfg.Password, "password", "p", dcfg.Password, "Password")
	cmd.Flags().StringVarP(&secretName, "secret", "s", "", "Secret name (required)")
	cmd.Flags().StringVar(&secretType, "type", "", "Type of secret: binary, card, credentials (required)")
	cmd.Flags().StringVar(&secretValue, "value", "", "Secret value (file path, card string, or credentials)")

	// Mark required flags
	_ = cmd.MarkFlagRequired("secret")
	_ = cmd.MarkFlagRequired("type")
	_ = cmd.MarkFlagRequired("value")

	return cmd
}
