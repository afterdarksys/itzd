package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"itz.agency/itzd/internal/config"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with itz.agency using your wallet",
	Long: `Login using your Ethereum wallet.

Steps:
  1. itzd requests a sign challenge from itz.agency
  2. You sign the message with your wallet (MetaMask, cast, etc.)
  3. itzd submits the signature and stores the JWT token

For programmatic use, set ITZ_TOKEN env var instead of interactive login.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg := loadConfig()
		api := newClient(cfg)

		reader := bufio.NewReader(os.Stdin)

		// Get wallet address
		fmt.Print("Wallet address (0x...): ")
		address, _ := reader.ReadString('\n')
		address = strings.TrimSpace(address)
		if address == "" {
			fmt.Fprintln(os.Stderr, "error: wallet address required")
			os.Exit(1)
		}

		// Get challenge
		fmt.Printf("Requesting challenge from %s...\n", cfg.APIEndpoint)
		challenge, err := api.Challenge(address, "ethereum")
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}

		// Show message to sign
		fmt.Println()
		fmt.Println("┌─ Message to sign ────────────────────────────────────────────")
		for _, line := range strings.Split(challenge.Message, "\n") {
			fmt.Println("│ " + line)
		}
		fmt.Println("└──────────────────────────────────────────────────────────────")
		fmt.Println()
		fmt.Println("Sign this message with your wallet and paste the signature below.")
		fmt.Println("Using cast:  cast wallet sign --account <account> --password '' \\")
		fmt.Printf("             \"$(itzd login --get-message %s)\"\n", address)
		fmt.Println()
		fmt.Print("Signature (0x...): ")
		signature, _ := reader.ReadString('\n')
		signature = strings.TrimSpace(signature)
		if signature == "" {
			fmt.Fprintln(os.Stderr, "error: signature required")
			os.Exit(1)
		}

		// Verify
		fmt.Println("Verifying signature...")
		auth, err := api.VerifyEthereum(address, challenge.Challenge, signature)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}

		// Save token
		cfg.Token = auth.AccessToken
		cfg.WalletAddr = address
		cfg.Network = "ethereum"
		if err := config.Save(cfg); err != nil {
			fmt.Fprintln(os.Stderr, "error saving token:", err)
			os.Exit(1)
		}

		fmt.Println()
		fmt.Printf("✓ Logged in as %s\n", auth.Address)
		fmt.Printf("  User ID: %s\n", auth.UserID)
		fmt.Printf("  Token stored in ~/.itzd/config.json\n")
	},
}
