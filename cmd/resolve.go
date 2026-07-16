package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"itz.agency/itzd/internal/config"
	"itz.agency/itzd/internal/resolver"
)

type resolveService interface {
	Resolve(context.Context, string, string, resolver.Policy) (*resolver.Result, error)
}

type resolverBuilder func() (resolveService, error)

func buildDNSResolver() (resolveService, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}
	transport, err := resolver.NewDoTTransport(resolver.DoTConfig{
		Endpoint: cfg.ValidatorEndpoint, ServerName: cfg.ValidatorServerName, Timeout: 5 * time.Second,
	})
	if err != nil {
		return nil, err
	}
	return resolver.New(transport, resolver.DefaultValidatorRegistry(), cfg.DefaultZone, time.Now), nil
}

func newResolveCommand(build resolverBuilder) *cobra.Command {
	var chain, network string
	var jsonOutput, allowInsecure bool
	command := &cobra.Command{
		Use:          "resolve <name> --chain <caip-2>",
		Short:        "Resolve a wallet address through DNSSEC-secure DNS",
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if network != "" {
				aliases := map[string]string{"e": "eip155:1", "b": "bip122:000000000019d6689c085ae165831e93", "s": "solana:mainnet"}
				mapped, ok := aliases[network]
				if !ok {
					return fmt.Errorf("unknown deprecated network alias %q", network)
				}
				if chain != "" && chain != mapped {
					return fmt.Errorf("--chain conflicts with --network")
				}
				chain = mapped
				cmd.PrintErrln("warning: --network is deprecated; use --chain " + mapped)
			}
			if chain == "" {
				return fmt.Errorf("--chain is required")
			}
			policy := resolver.StrictPolicy()
			if allowInsecure {
				policy = resolver.InsecureDevelopmentPolicy()
				cmd.PrintErrln("WARNING: --allow-insecure disables the DNSSEC security requirement; development use only")
			}
			service, err := build()
			if err != nil {
				return fmt.Errorf("initializing resolver: %w", err)
			}
			result, err := service.Resolve(cmd.Context(), args[0], chain, policy)
			if err != nil {
				return err
			}
			if jsonOutput {
				encoder := json.NewEncoder(cmd.OutOrStdout())
				encoder.SetIndent("", "  ")
				return encoder.Encode(result)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%s  %s\n", result.Chain, result.Address)
			fmt.Fprintf(cmd.OutOrStdout(), "DNS: %s\nDNSSEC: %s via %s\nTTL: %ds\n", result.QueryName, result.DNSSEC, result.Validator, result.TTL)
			return nil
		},
	}
	command.Flags().StringVar(&chain, "chain", "", "Exact CAIP-2 chain identifier (required)")
	command.Flags().StringVarP(&network, "network", "n", "", "Deprecated short network alias")
	command.Flags().BoolVar(&jsonOutput, "json", false, "Emit structured JSON")
	command.Flags().BoolVar(&allowInsecure, "allow-insecure", false, "Allow insecure DNS in development")
	return command
}

var resolveCmd = newResolveCommand(buildDNSResolver)

func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}
