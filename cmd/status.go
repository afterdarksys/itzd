package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check itz.agency API status",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := loadConfig()
		api := newClient(cfg)

		health, err := api.Health()
		if err != nil {
			fmt.Fprintf(os.Stderr, "✗ API unreachable: %s (%s)\n", cfg.APIEndpoint, err)
			os.Exit(1)
		}

		fmt.Printf("✓ API online: %s\n", cfg.APIEndpoint)
		if env, ok := health["environment"].(string); ok {
			fmt.Printf("  Environment : %s\n", env)
		}
		if ver, ok := health["version"].(string); ok {
			fmt.Printf("  Version     : %s\n", ver)
		}
		if ts, ok := health["timestamp"].(string); ok {
			fmt.Printf("  Timestamp   : %s\n", ts)
		}

		if cfg.Token != "" {
			fmt.Printf("  Auth        : logged in (%s)\n", cfg.WalletAddr)
		} else {
			fmt.Printf("  Auth        : not logged in\n")
		}
	},
}
