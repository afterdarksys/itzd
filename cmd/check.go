package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var checkCmd = &cobra.Command{
	Use:   "check <username> <network>",
	Short: "Check if a username is available",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		username := strings.ToLower(strings.TrimSpace(args[0]))
		network := strings.ToLower(strings.TrimSpace(args[1]))

		cfg := loadConfig()
		api := newClient(cfg)

		result, err := api.Check(username, network)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}

		available, _ := result["available"].(bool)
		if available {
			fmt.Printf("✓ %s.itz.agency [%s] is available\n", username, network)
		} else {
			reason, _ := result["reason"].(string)
			fmt.Printf("✗ %s.itz.agency [%s] is taken", username, network)
			if reason != "" {
				fmt.Printf(" — %s", reason)
			}
			fmt.Println()
		}
	},
}
