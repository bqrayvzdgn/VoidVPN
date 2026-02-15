package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/voidvpn/voidvpn/internal/config"
	"github.com/voidvpn/voidvpn/internal/keystore"
	"github.com/voidvpn/voidvpn/internal/ui"
)

var serversCmd = &cobra.Command{
	Use:   "servers",
	Short: "Manage VPN servers",
}

var serversListCmd = &cobra.Command{
	Use:   "list",
	Short: "List configured servers",
	RunE: func(cmd *cobra.Command, args []string) error {
		servers, err := config.ListServers()
		if err != nil {
			return fmt.Errorf("failed to list servers: %w", err)
		}

		fmt.Println(ui.TitleStyle.Render("Configured Servers"))
		fmt.Println()

		columns := []ui.TableColumn{
			{Header: "Name", Width: 20},
			{Header: "Proto", Width: 6},
			{Header: "Endpoint", Width: 30},
			{Header: "Address", Width: 18},
			{Header: "DNS", Width: 20},
		}

		var rows []ui.TableRow
		for _, s := range servers {
			dns := ""
			if len(s.DNS) > 0 {
				dns = s.DNS[0]
				if len(s.DNS) > 1 {
					dns += fmt.Sprintf(" (+%d)", len(s.DNS)-1)
				}
			}
			proto := protocolLabel(s.Protocol)
			rows = append(rows, ui.TableRow{s.Name, proto, s.Endpoint, s.Address, dns})
		}

		fmt.Println(ui.RenderTable(columns, rows))
		return nil
	},
}

var serversAddCmd = &cobra.Command{
	Use:   "add <name>",
	Short: "Add a server configuration",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		if config.ServerExists(name) {
			return fmt.Errorf("server '%s' already exists. Remove it first or choose a different name", name)
		}

		endpoint, _ := cmd.Flags().GetString("endpoint")
		publicKey, _ := cmd.Flags().GetString("public-key")
		address, _ := cmd.Flags().GetString("address")
		dns, _ := cmd.Flags().GetStringSlice("dns")

		if endpoint == "" || publicKey == "" || address == "" {
			return fmt.Errorf("required flags: --endpoint, --public-key, --address")
		}

		server := config.DefaultServerConfig()
		server.Name = name
		server.Endpoint = endpoint
		server.PublicKey = publicKey
		server.Address = address
		if len(dns) > 0 {
			server.DNS = dns
		}

		if err := config.SaveServer(server); err != nil {
			return fmt.Errorf("failed to save server: %w", err)
		}

		fmt.Println(ui.SuccessStyle.Render(fmt.Sprintf("✓ Server '%s' added", name)))
		return nil
	},
}

var serversRemoveCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove a server configuration",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		if err := config.RemoveServer(name); err != nil {
			return err
		}

		fmt.Println(ui.SuccessStyle.Render(fmt.Sprintf("✓ Server '%s' removed", name)))
		return nil
	},
}

var importName string

var serversImportCmd = &cobra.Command{
	Use:   "import <file>",
	Short: "Import a WireGuard (.conf) or OpenVPN (.ovpn) configuration file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := args[0]

		ext := strings.ToLower(path)
		switch {
		case strings.HasSuffix(ext, ".ovpn"):
			return importOpenVPN(path)
		default:
			return importWireGuard(path)
		}
	},
}

func importWireGuard(path string) error {
	server, privateKey, err := config.ImportWireGuardConfig(path)
	if err != nil {
		return fmt.Errorf("failed to import config: %w", err)
	}

	if importName != "" {
		server.Name = importName
	}

	if config.ServerExists(server.Name) {
		return fmt.Errorf("server '%s' already exists. Remove it first or rename the config file", server.Name)
	}

	if err := config.SaveServer(server); err != nil {
		return fmt.Errorf("failed to save server: %w", err)
	}

	if privateKey != "" {
		ks := keystore.New()
		if err := ks.Store(server.Name, privateKey); err != nil {
			fmt.Println(ui.WarningStyle.Render(fmt.Sprintf("⚠ Failed to store private key: %v", err)))
			fmt.Println(ui.DimStyle.Render("  You can manually store it with 'voidvpn keygen --save'"))
		} else {
			fmt.Println(ui.SuccessStyle.Render("✓ Private key stored in keystore"))
		}
	}

	fmt.Println(ui.SuccessStyle.Render(fmt.Sprintf("✓ Imported WireGuard server '%s'", server.Name)))
	fmt.Printf("  %s %s\n", ui.LabelStyle.Render("Endpoint:"), server.Endpoint)
	fmt.Printf("  %s %s\n", ui.LabelStyle.Render("Address:"), server.Address)
	fmt.Printf("  %s %s\n", ui.LabelStyle.Render("DNS:"), fmt.Sprintf("%v", server.DNS))

	return nil
}

func importOpenVPN(path string) error {
	server, err := config.ImportOpenVPNConfig(path)
	if err != nil {
		return fmt.Errorf("failed to import config: %w", err)
	}

	if importName != "" {
		server.Name = importName
	}

	if config.ServerExists(server.Name) {
		return fmt.Errorf("server '%s' already exists. Remove it first or rename the config file", server.Name)
	}

	if err := config.SaveServer(server); err != nil {
		return fmt.Errorf("failed to save server: %w", err)
	}

	fmt.Println(ui.SuccessStyle.Render(fmt.Sprintf("✓ Imported OpenVPN server '%s'", server.Name)))
	fmt.Printf("  %s %s\n", ui.LabelStyle.Render("Endpoint:"), server.Endpoint)
	fmt.Printf("  %s %s\n", ui.LabelStyle.Render("Protocol:"), server.Proto)

	return nil
}

func protocolLabel(protocol string) string {
	switch protocol {
	case "openvpn":
		return "OVPN"
	default:
		return "WG"
	}
}

func init() {
	serversAddCmd.Flags().String("endpoint", "", "Server endpoint (host:port)")
	serversAddCmd.Flags().String("public-key", "", "Server's WireGuard public key")
	serversAddCmd.Flags().String("address", "", "Tunnel IP address (e.g., 10.0.0.2/24)")
	serversAddCmd.Flags().StringSlice("dns", nil, "DNS servers (comma-separated)")

	serversImportCmd.Flags().StringVar(&importName, "name", "", "Custom name for the imported server")

	serversCmd.AddCommand(serversListCmd)
	serversCmd.AddCommand(serversAddCmd)
	serversCmd.AddCommand(serversRemoveCmd)
	serversCmd.AddCommand(serversImportCmd)
}
