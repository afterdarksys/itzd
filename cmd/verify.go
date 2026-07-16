package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"itz.agency/itzd/internal/client"
	"itz.agency/itzd/internal/config"
	"itz.agency/itzd/internal/resolver"
)

type parityClient interface {
	ResolveNetwork(name, network string) (*client.WalletRecord, error)
}

type verifyDependencies struct {
	resolver resolveService
	api      parityClient
}
type verifyBuilder func() (verifyDependencies, error)

type verifyReport struct {
	*resolver.Result
	AddressValid bool   `json:"address_valid"`
	APIReachable bool   `json:"api_reachable"`
	APIParity    *bool  `json:"api_dns_parity,omitempty"`
	APIError     string `json:"api_error,omitempty"`
}

func buildVerifyDependencies() (verifyDependencies, error) {
	service, err := buildDNSResolver()
	if err != nil {
		return verifyDependencies{}, err
	}
	cfg, err := config.Load()
	if err != nil {
		return verifyDependencies{}, err
	}
	return verifyDependencies{resolver: service, api: client.New(cfg.APIEndpoint, cfg.Token)}, nil
}

func newVerifyCommand(build verifyBuilder) *cobra.Command {
	var chain string
	var jsonOutput, checkAPI bool
	command := &cobra.Command{
		Use: "verify <name> --chain <caip-2>", Short: "Diagnose secure DNS resolution and optional API parity",
		Args: cobra.ExactArgs(1), SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if chain == "" {
				return fmt.Errorf("--chain is required")
			}
			deps, err := build()
			if err != nil {
				return fmt.Errorf("initializing diagnostics: %w", err)
			}
			result, err := deps.resolver.Resolve(cmd.Context(), args[0], chain, resolver.StrictPolicy())
			if err != nil {
				return err
			}
			report := verifyReport{Result: result, AddressValid: true}
			mismatch := false
			if checkAPI {
				networks := map[string]string{"eip155:1": "e", "bip122:000000000019d6689c085ae165831e93": "b", "solana:mainnet": "s"}
				network, supported := networks[chain]
				name := strings.TrimSuffix(result.Name, ".itz.agency")
				if !supported || strings.Contains(name, ".") || deps.api == nil {
					report.APIError = "API parity is unavailable for this chain or custom domain"
				} else if apiRecord, apiErr := deps.api.ResolveNetwork(name, network); apiErr != nil {
					report.APIError = "API unavailable"
				} else {
					report.APIReachable = true
					parity := apiRecord.WalletAddress == result.Address
					report.APIParity = &parity
					mismatch = !parity
				}
			}
			if jsonOutput {
				encoder := json.NewEncoder(cmd.OutOrStdout())
				encoder.SetIndent("", "  ")
				if err := encoder.Encode(report); err != nil {
					return err
				}
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "DNSSEC: %s via %s\nAddress: %s (%s)\n", result.DNSSEC, result.Validator, result.Address, result.Chain)
				if checkAPI {
					if report.APIParity != nil {
						fmt.Fprintf(cmd.OutOrStdout(), "API parity: %t\n", *report.APIParity)
					} else {
						fmt.Fprintf(cmd.OutOrStdout(), "API parity: unavailable (%s)\n", report.APIError)
					}
				}
			}
			if mismatch {
				return &resolver.ResolveError{Code: resolver.CodeAPIDNSMismatch, Name: result.Name, Chain: chain, Reason: "secure DNS and API addresses differ"}
			}
			return nil
		},
	}
	command.Flags().StringVar(&chain, "chain", "", "Exact CAIP-2 chain identifier (required)")
	command.Flags().BoolVar(&jsonOutput, "json", false, "Emit structured JSON")
	command.Flags().BoolVar(&checkAPI, "check-api", false, "Compare secure DNS with the HTTP API")
	return command
}

var verifyCmd = newVerifyCommand(buildVerifyDependencies)
