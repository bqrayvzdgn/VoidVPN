package cli

import (
	"context"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/voidvpn/voidvpn/internal/config"
	"github.com/voidvpn/voidvpn/internal/daemon"
	"github.com/voidvpn/voidvpn/internal/keystore"
	"github.com/voidvpn/voidvpn/internal/logger"
	"github.com/voidvpn/voidvpn/internal/openvpn"
	"github.com/voidvpn/voidvpn/internal/platform"
	"github.com/voidvpn/voidvpn/internal/tunnel"
	"github.com/voidvpn/voidvpn/internal/ui"
	"github.com/voidvpn/voidvpn/internal/wireguard"
)

var (
	connectDaemon bool
)

var connectCmd = &cobra.Command{
	Use:   "connect [server]",
	Short: "Connect to a VPN server",
	Long:  "Connect to a configured VPN server (WireGuard or OpenVPN). Requires administrator/root privileges.",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Require admin/root — needed for network adapter configuration
		if !platform.IsAdmin() {
			return fmt.Errorf("administrator/root privileges required.\nOn Windows: right-click terminal and select 'Run as administrator'\nOn Linux/macOS: use 'sudo voidvpn connect'")
		}

		// Check if already connected
		if daemon.IsConnected() {
			return fmt.Errorf("already connected. Run 'voidvpn disconnect' first")
		}

		// Determine server name
		serverName := ""
		if len(args) > 0 {
			serverName = args[0]
		} else {
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}
			if cfg.DefaultServer == "" {
				return fmt.Errorf("no server specified and no default server configured.\nUsage: voidvpn connect <server>")
			}
			serverName = cfg.DefaultServer
		}

		// Load server config
		serverCfg, err := config.LoadServer(serverName)
		if err != nil {
			return fmt.Errorf("server '%s' not found. Run 'voidvpn servers list' to see available servers", serverName)
		}

		// Create the appropriate tunnel based on protocol
		var tun tunnel.Tunnel

		switch serverCfg.Protocol {
		case "openvpn":
			// OpenVPN uses the Interactive Service on Windows — no admin required
			tun = openvpn.NewTunnel(serverCfg)
		default:
			// WireGuard requires admin/root privileges
			if !platform.IsAdmin() {
				return fmt.Errorf("administrator/root privileges required for WireGuard.\nOn Windows: right-click terminal and select 'Run as administrator'\nOn Linux/macOS: use 'sudo voidvpn connect'")
			}
			ks := keystore.New()
			privateKey, err := ks.Load(serverName)
			if err != nil {
				// Try default key
				privateKey, err = ks.Load("default")
				if err != nil {
					return fmt.Errorf("no private key found for '%s'. Run 'voidvpn keygen --save' or import a config", serverName)
				}
			}
			tun = wireguard.NewTunnel(serverCfg, privateKey)
		}

		fmt.Println(ui.Banner())

		d := daemon.New(tun, serverCfg)

		// Pause logs before starting daemon to prevent interleaved output with spinner
		logger.Pause()

		// Run daemon in background
		connectErr := make(chan error, 1)
		go func() {
			ctx := context.Background()
			connectErr <- d.Run(ctx)
		}()

		// Show spinner while connecting
		spinnerModel := ui.NewSpinner(fmt.Sprintf("Connecting to %s...", serverName))
		p := tea.NewProgram(spinnerModel)

		go func() {
			select {
			case err := <-connectErr:
				// Run() returned early — connection failed
				p.Send(ui.ConnectMsg{Err: err})
			case <-d.Connected:
				// Tunnel connected successfully
				p.Send(ui.ConnectMsg{Err: nil})
			}
		}()

		if _, err := p.Run(); err != nil {
			logger.Resume()
			return fmt.Errorf("UI error: %w", err)
		}

		// Resume logs after spinner is done
		logger.Resume()

		// If connection succeeded, keep running until daemon exits (Ctrl+C)
		if err := <-connectErr; err != nil {
			return err
		}

		return nil
	},
}

func init() {
	connectCmd.Flags().BoolVar(&connectDaemon, "daemon", false, "Run in background (daemon mode)")

	// Suppress usage on the unused variable
	_ = os.Stdout
}
