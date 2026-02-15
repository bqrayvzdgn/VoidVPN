package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/voidvpn/voidvpn/internal/ui"
	"github.com/voidvpn/voidvpn/pkg/version"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(ui.Banner())
		fmt.Println(ui.AccentStyle.Render(version.Full()))
	},
}
