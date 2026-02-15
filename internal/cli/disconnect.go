package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/voidvpn/voidvpn/internal/daemon"
	"github.com/voidvpn/voidvpn/internal/ui"
)

var disconnectCmd = &cobra.Command{
	Use:   "disconnect",
	Short: "Disconnect from VPN",
	RunE: func(cmd *cobra.Command, args []string) error {
		if !daemon.IsConnected() {
			fmt.Println(ui.WarningStyle.Render("Not connected to any VPN server."))
			return nil
		}

		// Send disconnect via IPC
		resp, err := daemon.SendIPCRequest("disconnect")
		if err != nil {
			// If IPC fails, try to clean up state file directly
			if cleanErr := daemon.ClearState(); cleanErr != nil {
				return fmt.Errorf("failed to disconnect: %w", err)
			}
			fmt.Println(ui.WarningStyle.Render("Connection state cleared (daemon may have already exited)."))
			return nil
		}

		if !resp.Success {
			return fmt.Errorf("disconnect failed: %s", resp.Error)
		}

		fmt.Println(ui.SuccessStyle.Render("✓ Disconnected from VPN"))
		return nil
	},
}
