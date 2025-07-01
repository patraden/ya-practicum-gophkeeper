package cmd

import (
	"github.com/patraden/ya-practicum-gophkeeper/client/internal/app"
	"github.com/patraden/ya-practicum-gophkeeper/client/internal/config"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/logger"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

func NewSyncCmd(dcfg *config.Config) *cobra.Command {
	log := logger.StdoutConsole(zerolog.DebugLevel)

	var secretName string

	cmd := &cobra.Command{
		Use:   "sync",
		Short: "syncronizes user's secret witth gophkeeper server",
		RunE: func(_ *cobra.Command, _ []string) error {
			cfg := config.LoadConfig(dcfg)
			return app.SyncSecrets(cfg, secretName, log)
		},
		SilenceUsage: true,
	}

	cmd.Flags().StringVarP(&dcfg.Username, "username", "u", dcfg.Username, "Username (required)")
	cmd.Flags().StringVarP(&dcfg.Password, "password", "p", dcfg.Password, "Password (required)")
	cmd.Flags().StringVarP(&secretName, "secret", "s", "", "Secret name (required)")
	_ = cmd.MarkFlagRequired("secret")

	return cmd
}
