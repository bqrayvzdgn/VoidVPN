package cli

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/voidvpn/voidvpn/internal/daemon"
	"github.com/voidvpn/voidvpn/internal/ui"
)

var (
	statusWatch bool
	statusJSON  bool
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show connection status",
	RunE: func(cmd *cobra.Command, args []string) error {
		if statusWatch {
			return watchStatus()
		}
		return showStatus()
	},
}

func showStatus() error {
	if !daemon.IsConnected() {
		info := ui.StatusInfo{Connected: false}
		if statusJSON {
			data, _ := json.MarshalIndent(map[string]interface{}{
				"connected": false,
			}, "", "  ")
			fmt.Println(string(data))
			return nil
		}
		fmt.Print(ui.RenderStatus(info))
		return nil
	}

	// Get live status via IPC
	resp, err := daemon.SendIPCRequest("status")
	if err != nil {
		// Fall back to state file
		state, stateErr := daemon.LoadState()
		if stateErr != nil {
			return fmt.Errorf("failed to get status: %w", err)
		}
		return renderState(state)
	}

	if !resp.Success {
		return fmt.Errorf("failed to get status: %s", resp.Error)
	}

	return renderState(resp.State)
}

func renderState(state *daemon.ConnectionState) error {
	if statusJSON {
		data, err := json.MarshalIndent(state, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	}

	info := ui.StatusInfo{
		Connected:   true,
		Protocol:    state.Protocol,
		ServerName:  state.Server,
		Endpoint:    state.Endpoint,
		TunnelIP:    state.TunnelIP,
		ConnectedAt: state.ConnectedAt,
		TxBytes:     state.TxBytes,
		RxBytes:     state.RxBytes,
	}
	fmt.Print(ui.RenderStatus(info))
	return nil
}

func watchStatus() error {
	for {
		// Clear screen
		fmt.Print("\033[H\033[2J")

		if err := showStatus(); err != nil {
			fmt.Println(ui.ErrorStyle.Render(err.Error()))
		}

		fmt.Println()
		fmt.Println(ui.DimStyle.Render("Refreshing every 2s... Press Ctrl+C to stop."))

		time.Sleep(2 * time.Second)
	}
}

func init() {
	statusCmd.Flags().BoolVar(&statusWatch, "watch", false, "Live-updating status display")
	statusCmd.Flags().BoolVar(&statusJSON, "json", false, "Output status as JSON")
}
