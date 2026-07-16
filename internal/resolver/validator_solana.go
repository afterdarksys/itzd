package resolver

import (
	"fmt"

	"github.com/mr-tron/base58"
)

type SolanaValidator struct{}

func (SolanaValidator) Validate(address string) error {
	decoded, err := base58.Decode(address)
	if err != nil {
		return fmt.Errorf("invalid Solana base58 address")
	}
	if len(decoded) != 32 {
		return fmt.Errorf("Solana address must decode to 32 bytes")
	}
	return nil
}
