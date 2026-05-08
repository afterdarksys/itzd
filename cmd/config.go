package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"itz.agency/itzd/internal/config"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Show or set itzd configuration",
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current config",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := loadConfig()
		fmt.Printf("API endpoint : %s\n", cfg.APIEndpoint)
		if cfg.WalletAddr != "" {
			fmt.Printf("Wallet       : %s\n", cfg.WalletAddr)
			fmt.Printf("Network      : %s\n", cfg.Network)
		}
		if cfg.Token != "" {
			fmt.Printf("Auth         : logged in\n")
		} else {
			fmt.Printf("Auth         : not logged in\n")
		}
	},
}

var configSetAPICmd = &cobra.Command{
	Use:   "set-api <url>",
	Short: "Set the API endpoint",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := loadConfig()
		cfg.APIEndpoint = args[0]
		if err := config.Save(cfg); err != nil {
			fmt.Println("error saving config:", err)
			return
		}
		fmt.Println("API endpoint set to", args[0])
	},
}

var configResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset config (clears token and wallet)",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := &config.Config{APIEndpoint: config.DefaultAPIEndpoint}
		if err := config.Save(cfg); err != nil {
			fmt.Println("error resetting config:", err)
			return
		}
		fmt.Println("Config reset.")
	},
}

func init() {
	configCmd.AddCommand(configShowCmd, configSetAPICmd, configResetCmd)
}
