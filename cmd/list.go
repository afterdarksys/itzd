package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List your registered usernames",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := loadConfig()
		requireAuth(cfg)
		api := newClient(cfg)

		names, err := api.MyUsernames()
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}

		if len(names) == 0 {
			fmt.Println("No usernames registered yet.")
			fmt.Println("Run: itzd register <username> <network> <wallet>")
			return
		}

		fmt.Printf("Your usernames (%d)\n", len(names))
		fmt.Println(strings.Repeat("─", 60))
		for _, n := range names {
			verified := ""
			if n.Verified {
				verified = "  ✓ verified"
			}
			synced := ""
			if !n.DNSSynced {
				synced = "  ⚠ dns pending"
			}
			fmt.Printf("  %s.itz.agency  [%s]  %s%s%s\n",
				n.Username, n.Network, n.WalletAddress, verified, synced)
		}
	},
}
