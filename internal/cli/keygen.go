package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/voidvpn/voidvpn/internal/keystore"
	"github.com/voidvpn/voidvpn/internal/ui"
	"github.com/voidvpn/voidvpn/internal/wireguard"
)

var (
	keygenSave bool
	keygenName string
)

var keygenCmd = &cobra.Command{
	Use:   "keygen",
	Short: "Generate a WireGuard keypair",
	RunE: func(cmd *cobra.Command, args []string) error {
		kp, err := wireguard.GenerateKeyPair()
		if err != nil {
			return fmt.Errorf("failed to generate keypair: %w", err)
		}

		fmt.Println(ui.TitleStyle.Render("WireGuard Keypair Generated"))
		fmt.Println()
		fmt.Printf("%s %s\n", ui.LabelStyle.Render("Private Key:"), ui.DimStyle.Render(kp.PrivateKey))
		fmt.Printf("%s %s\n", ui.LabelStyle.Render("Public Key:"), ui.AccentStyle.Render(kp.PublicKey))

		if keygenSave {
			name := keygenName
			if name == "" {
				name = "default"
			}

			ks := keystore.New()
			if err := ks.Store(name, kp.PrivateKey); err != nil {
				return fmt.Errorf("failed to store private key: %w", err)
			}

			fmt.Println()
			fmt.Println(ui.SuccessStyle.Render(fmt.Sprintf("✓ Private key saved to keystore as '%s'", name)))
		} else {
			fmt.Println()
			fmt.Println(ui.WarningStyle.Render("⚠ Save your private key securely! Use --save to store in OS keyring."))
		}

		return nil
	},
}

func init() {
	keygenCmd.Flags().BoolVar(&keygenSave, "save", false, "Save private key to OS keyring")
	keygenCmd.Flags().StringVar(&keygenName, "name", "default", "Name for the stored key")
}
