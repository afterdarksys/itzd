package resolver

import "testing"

func TestAddressValidators(t *testing.T) {
	tests := []struct {
		name, chain, address string
		valid                bool
	}{
		{"EVM lowercase", "eip155:1", "0x0000000000000000000000000000000000000001", true},
		{"EVM checksummed", "eip155:1", "0x52908400098527886E0F7030069857D2E4169EE7", true},
		{"EVM short", "eip155:1", "0x1234", false},
		{"EVM non-hex", "eip155:1", "0xgg00000000000000000000000000000000000000", false},
		{"EVM bad mixed checksum", "eip155:1", "0x52908400098527886e0F7030069857D2E4169EE7", false},
		{"Solana system", "solana:mainnet", "11111111111111111111111111111111", true},
		{"Solana short", "solana:mainnet", "1111", false},
		{"Solana invalid base58", "solana:mainnet", "00000000000000000000000000000000", false},
		{"Bitcoin mainnet", "bip122:000000000019d6689c085ae165831e93", "1BoatSLRHtKNngkdXEeobR76b53LETtpyT", true},
		{"Bitcoin testnet rejected", "bip122:000000000019d6689c085ae165831e93", "mipcBbFg9gMiCh81Kj8tqqdgoZub1ZJRfn", false},
	}
	registry := DefaultValidatorRegistry()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := registry.Validate(tt.chain, tt.address)
			if (err == nil) != tt.valid {
				t.Fatalf("valid=%v err=%v", tt.valid, err)
			}
		})
	}
}

func TestAddressValidatorRejectsUnsupportedChain(t *testing.T) {
	err := DefaultValidatorRegistry().Validate("example:network", "anything")
	resolveErr, ok := err.(*ResolveError)
	if !ok || resolveErr.Code != CodeInvalidAddress {
		t.Fatalf("got %v", err)
	}
}
