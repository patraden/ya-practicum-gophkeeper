package cmd

import (
	"github.com/patraden/ya-practicum-gophkeeper/client/internal/app"
	"github.com/patraden/ya-practicum-gophkeeper/client/internal/config"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/logger"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

func NewRegisterCmd(dcfg *config.Config) *cobra.Command {
	log := logger.StdoutConsole(zerolog.DebugLevel)
	cmd := &cobra.Command{
		Use:   "register",
		Short: "Register new user in gophkeeper",
		RunE: func(_ *cobra.Command, _ []string) error {
			cfg := config.LoadConfig(dcfg)
			return app.RegisterUser(cfg, log)
		},
		SilenceUsage: true,
	}

	cmd.Flags().StringVarP(&dcfg.Username, "username", "u", dcfg.Username, "Username (required)")
	cmd.Flags().StringVarP(&dcfg.Password, "password", "p", dcfg.Password, "Password (required)")

	return cmd
}
