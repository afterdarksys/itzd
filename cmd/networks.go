package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var networksCmd = &cobra.Command{
	Use:   "networks",
	Short: "List all supported blockchain networks",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Supported networks")
		fmt.Println("──────────────────────────────────────────────────────────────")
		fmt.Println("  Code   Chain                      CAIP-2 Namespace")
		fmt.Println("──────────────────────────────────────────────────────────────")

		networks := []struct{ code, name, caip2 string }{
			// EVM
			{"e",    "Ethereum mainnet",          "eip155:1"},
			{"poly", "Polygon",                   "eip155:137"},
			{"arb",  "Arbitrum One",              "eip155:42161"},
			{"op",   "Optimism",                  "eip155:10"},
			{"cb",   "Base (Coinbase L2)",         "eip155:8453"},
			{"bnb",  "BNB Chain",                 "eip155:56"},
			{"a",    "Avalanche C-Chain",          "eip155:43114"},
			{"zk",   "zkSync Era",                "eip155:324"},
			{"ftm",  "Fantom",                    "eip155:250"},
			{"celo", "Celo",                      "eip155:42220"},
			// Non-EVM
			{"b",    "Bitcoin",                   "bip122:btc"},
			{"s",    "Solana",                    "solana:sol"},
			{"apt",  "Aptos",                     "aptos:mainnet"},
			{"r",    "Ripple/XRP",                "xrpl:mainnet"},
			{"near", "NEAR Protocol",             "near:mainnet"},
			{"sui",  "Sui",                       "sui:mainnet"},
			{"algo", "Algorand",                  "algorand:mainnet"},
			// Cosmos
			{"atom", "Cosmos Hub",                "cosmos:cosmoshub-4"},
			{"osmo", "Osmosis",                   "cosmos:osmosis-1"},
			// Polkadot
			{"dot",  "Polkadot",                  "polkadot:91b171bb"},
			{"ksm",  "Kusama",                    "polkadot:b0a8d493"},
			// Other
			{"xtz",  "Tezos",                     "tezos:mainnet"},
			{"ada",  "Cardano",                   "cardano:mainnet"},
			{"hbar", "Hedera Hashgraph",          "hedera:mainnet"},
		}

		for _, n := range networks {
			fmt.Printf("  %-6s %-26s %s\n", n.code, n.name, n.caip2)
		}
		fmt.Println()
		fmt.Println("Usage: itzd register <username> <code> <wallet>")
		fmt.Println("       itzd resolve <username> --network <code>")
	},
}
