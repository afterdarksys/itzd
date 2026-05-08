package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"itz.agency/itzd/internal/client"
)

var resolveNetwork string

var resolveCmd = &cobra.Command{
	Use:   "resolve <username>",
	Short: "Resolve a username to wallet address(es)",
	Long: `Resolve a username registered on itz.agency to its wallet address(es).

Examples:
  itzd resolve ryan              → all wallets for 'ryan'
  itzd resolve ryan --network e  → Ethereum wallet only
  itzd resolve ryan -n s         → Solana wallet only`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		username := strings.ToLower(strings.TrimSuffix(strings.TrimSpace(args[0]), ".itz.agency"))

		cfg := loadConfig()
		api := newClient(cfg)

		if resolveNetwork != "" {
			rec, err := api.ResolveNetwork(username, resolveNetwork)
			if err != nil {
				fmt.Fprintln(os.Stderr, "error:", err)
				os.Exit(1)
			}
			printRecord(rec)
			return
		}

		resp, err := api.Resolve(username)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}

		fmt.Printf("%s.itz.agency  (%d wallet%s)\n", resp.Username, resp.Count, plural(resp.Count))
		fmt.Println(strings.Repeat("─", 60))
		for i := range resp.Wallets {
			printRecord(&resp.Wallets[i])
		}
	},
}

func printRecord(w *client.WalletRecord) {
	verified := ""
	if w.Verified {
		verified = "  ✓ verified"
	}
	synced := ""
	if !w.DNSSynced {
		synced = "  ⚠ dns pending"
	}
	fmt.Printf("  %-6s  %s%s%s\n", w.Network, w.WalletAddress, verified, synced)
	fmt.Printf("         DNS: %s\n", w.DNSName)
	fmt.Printf("         TXT: %s\n", w.TXTRecord)
}

func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}

func init() {
	resolveCmd.Flags().StringVarP(&resolveNetwork, "network", "n", "", "Network code (e, b, s, poly, arb, ...)")
}
