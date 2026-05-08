package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var registerCmd = &cobra.Command{
	Use:   "register <username> <network> <wallet-address>",
	Short: "Register a username → wallet address mapping",
	Long: `Register a human-readable username and link it to your wallet address.

The registration publishes a DNS TXT record:
  _waddr.<username>.itz.agency  TXT  "<namespace>:<wallet>"

This record follows draft-chins-dnsop-web3-wallet-mapping-03 (IETF).

Examples:
  itzd register ryan e 0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb
  itzd register ryan b bc1qxy2kgdygjrsqtzq2n0yrf2493p83kkfjhx0wlh
  itzd register ryan s DRpbCBMxVnDK7maPM5tGv6MvB3v1sRMC86PZ8okm21hy

Network codes: e=ETH  b=BTC  s=SOL  poly=Polygon  arb=Arbitrum
               op=Optimism  cb=Base  atom=Cosmos  dot=Polkadot
               (run 'itzd networks' for full list)`,
	Args: cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		username := strings.ToLower(strings.TrimSpace(args[0]))
		network := strings.ToLower(strings.TrimSpace(args[1]))
		wallet := strings.TrimSpace(args[2])

		cfg := loadConfig()
		requireAuth(cfg)
		api := newClient(cfg)

		fmt.Printf("Registering %s.itz.agency [%s] → %s\n", username, network, wallet)

		rec, err := api.Register(username, network, wallet)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}

		fmt.Println()
		fmt.Printf("✓ Registered: %s.itz.agency\n", rec.Username)
		fmt.Printf("  Network  : %s (%s)\n", rec.Network, rec.Namespace)
		fmt.Printf("  Wallet   : %s\n", rec.WalletAddress)
		fmt.Printf("  DNS name : %s\n", rec.DNSName)
		if rec.DNSSynced {
			fmt.Printf("  DNS sync : ✓ live\n")
		} else {
			fmt.Printf("  DNS sync : ⚠ pending (may take up to 60s)\n")
		}
		fmt.Println()
		fmt.Println("Verify with:")
		fmt.Printf("  dig TXT %s\n", rec.DNSName)
		fmt.Printf("  itzd resolve %s\n", rec.Username)
	},
}
