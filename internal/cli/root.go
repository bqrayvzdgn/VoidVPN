// Package cli defines all Cobra commands for the VoidVPN CLI.
package cli

import (
	"github.com/spf13/cobra"
	"github.com/voidvpn/voidvpn/internal/config"
	"github.com/voidvpn/voidvpn/internal/logger"
)

var (
	verbose bool
)

var rootCmd = &cobra.Command{
	Use:   "voidvpn",
	Short: "VoidVPN — WireGuard VPN Client",
	Long:  "A fast, secure WireGuard VPN client. No external dependencies required.",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		_ = config.EnsureDirs()

		// Initialize logger from config, override with --verbose
		logLevel := "info"
		if cfg, err := config.Load(); err == nil && cfg.LogLevel != "" {
			logLevel = cfg.LogLevel
		}
		if verbose {
			logLevel = "debug"
		}
		logger.Init(logLevel)
	},
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")

	rootCmd.AddCommand(connectCmd)
	rootCmd.AddCommand(disconnectCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(serversCmd)
	rootCmd.AddCommand(keygenCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(versionCmd)
}

func Execute() error {
	return rootCmd.Execute()
}
