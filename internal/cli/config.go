package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/voidvpn/voidvpn/internal/config"
	"github.com/voidvpn/voidvpn/internal/ui"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage VoidVPN configuration",
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		fmt.Println(ui.TitleStyle.Render("VoidVPN Configuration"))
		fmt.Println()
		fmt.Printf("%s %s\n", ui.LabelStyle.Render("Log Level:"), ui.ValueStyle.Render(cfg.LogLevel))
		fmt.Printf("%s %s\n", ui.LabelStyle.Render("Default Server:"), ui.ValueStyle.Render(cfg.DefaultServer))
		fmt.Printf("%s %s\n", ui.LabelStyle.Render("Auto Connect:"), ui.ValueStyle.Render(fmt.Sprintf("%v", cfg.AutoConnect)))
		fmt.Printf("%s %s\n", ui.LabelStyle.Render("Kill Switch:"), ui.ValueStyle.Render(fmt.Sprintf("%v", cfg.KillSwitch)))
		fmt.Printf("%s %s\n", ui.LabelStyle.Render("DNS Fallback:"), ui.ValueStyle.Render(fmt.Sprintf("%v", cfg.DNSFallback)))
		fmt.Printf("%s %s\n", ui.LabelStyle.Render("Config Path:"), ui.DimStyle.Render(config.ConfigFile()))

		return nil
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Long: `Set a configuration value. Available keys:
  log_level       - Logging level (debug, info, warn, error)
  default_server  - Default server for quick connect
  auto_connect    - Auto-connect on startup (true/false)
  kill_switch     - Block traffic if VPN drops (true/false)`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key, value := args[0], args[1]

		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		if !cfg.Set(key, value) {
			return fmt.Errorf("unknown config key: %s", key)
		}

		if err := cfg.Save(); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		fmt.Println(ui.SuccessStyle.Render(fmt.Sprintf("✓ Set %s = %s", key, value)))
		return nil
	},
}

func init() {
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configSetCmd)
}
