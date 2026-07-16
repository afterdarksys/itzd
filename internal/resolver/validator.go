package resolver

type AddressValidator interface{ Validate(address string) error }

type ValidatorRegistry map[string]AddressValidator

func DefaultValidatorRegistry() ValidatorRegistry {
	evm := EVMValidator{}
	solana := SolanaValidator{}
	bitcoin := BitcoinMainnetValidator{}
	return ValidatorRegistry{
		"eip155:1":       evm,
		"solana:mainnet": solana,
		"solana:sol":     solana,
		"bip122:000000000019d6689c085ae165831e93": bitcoin,
		"bip122:btc": bitcoin,
	}
}

func (r ValidatorRegistry) Validate(chain, address string) error {
	validator, ok := r[chain]
	if !ok {
		return &ResolveError{Code: CodeInvalidAddress, Chain: chain, Reason: "address validation is not implemented for this chain"}
	}
	if err := validator.Validate(address); err != nil {
		return &ResolveError{Code: CodeInvalidAddress, Chain: chain, Reason: err.Error(), Cause: err}
	}
	return nil
}
