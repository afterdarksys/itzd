package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var whoamiCmd = &cobra.Command{
	Use:   "whoami <wallet-address>",
	Short: "Reverse lookup: find usernames registered to a wallet address",
	Long: `Find all itz.agency usernames associated with a wallet address.

Examples:
  itzd whoami 0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb
  itzd whoami DRpbCBMxVnDK7maPM5tGv6MvB3v1sRMC86PZ8okm21hy`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		address := strings.TrimSpace(args[0])

		cfg := loadConfig()
		api := newClient(cfg)

		resp, err := api.Whoami(address)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}

		fmt.Printf("%s  (%d username%s)\n", address, resp.Count, plural(resp.Count))
		fmt.Println(strings.Repeat("─", 60))
		for _, u := range resp.Usernames {
			verified := ""
			if u.Verified {
				verified = "  ✓"
			}
			fmt.Printf("  %s.itz.agency  [%s]%s\n", u.Username, u.Network, verified)
		}
	},
}
