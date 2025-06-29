package cmd

import (
	"github.com/patraden/ya-practicum-gophkeeper/client/internal/app"
	"github.com/patraden/ya-practicum-gophkeeper/client/internal/config"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/logger"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

func NewInstallCmd(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Installs gophkeeper cli",
		RunE: func(_ *cobra.Command, _ []string) error {
			log := logger.StdoutConsole(zerolog.DebugLevel)
			return app.SetupLocal(cfg, log)
		},
	}

	cmd.Flags().IntVarP(&cfg.ServerPort, "server-port", "p", cfg.ServerPort, "Server port")
	cmd.Flags().StringVarP(&cfg.ServerHost, "server-host", "a", cfg.ServerHost, "Server host")
	cmd.Flags().StringVarP(&cfg.ServerTLSCertPath, "server-ca-cert", "c", cfg.ServerTLSCertPath, "CA certificate path")
	cmd.Flags().StringVarP(&cfg.InstallDir, "dir", "d", cfg.InstallDir, "installation path")
	_ = cmd.MarkFlagRequired("path")

	return cmd
}
