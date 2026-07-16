// Package cmd implements the itzd CLI.
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"itz.agency/itzd/internal/client"
	"itz.agency/itzd/internal/config"
)

var rootCmd = &cobra.Command{
	Use:   "itzd",
	Short: "Fail-closed DNSSEC cryptocurrency address resolver",
	Long: `itzd resolves _waddr TXT records through an authenticated validating
DNS-over-TLS resolver. Secure DNS is the source of truth; the HTTP API is used
only for explicit diagnostics and never as an automatic fallback.`,
	SilenceErrors: true,
	SilenceUsage:  true,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// loadConfig loads the config and exits on failure.
func loadConfig() *config.Config {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error loading config:", err)
		os.Exit(1)
	}
	return cfg
}

// newClient builds an API client from config.
func newClient(cfg *config.Config) *client.Client {
	return client.New(cfg.APIEndpoint, cfg.Token)
}

// requireAuth checks that a token is stored in config.
func requireAuth(cfg *config.Config) {
	if cfg.Token == "" {
		fmt.Fprintln(os.Stderr, `not logged in — run: itzd login`)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(
		configCmd,
		loginCmd,
		resolveCmd,
		verifyCmd,
		whoamiCmd,
		registerCmd,
		checkCmd,
		listCmd,
		networksCmd,
		statusCmd,
	)
}
