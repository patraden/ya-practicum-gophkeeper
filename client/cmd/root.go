package cmd

import (
	"github.com/patraden/ya-practicum-gophkeeper/client/internal/config"
	"github.com/spf13/cobra"
)

func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gkcli",
		Short: "GophKeeper CLI",
	}

	cmd.SilenceErrors = true

	dcfg := config.DefaultConfig()
	cmd.PersistentFlags().BoolVar(&dcfg.DebugMode, "debug", dcfg.DebugMode, "Enable debug mode")

	cmd.AddCommand(NewInstallCmd(dcfg))
	cmd.AddCommand(NewRegisterCmd(dcfg))

	return cmd
}
