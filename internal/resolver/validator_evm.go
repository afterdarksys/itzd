package resolver

import (
	"encoding/hex"
	"fmt"
	"strings"

	"golang.org/x/crypto/sha3"
)

type EVMValidator struct{}

func (EVMValidator) Validate(address string) error {
	if len(address) != 42 || !strings.HasPrefix(address, "0x") {
		return fmt.Errorf("EVM address must be 20-byte 0x hex")
	}
	body := address[2:]
	if _, err := hex.DecodeString(body); err != nil {
		return fmt.Errorf("EVM address contains non-hex characters")
	}
	lower, upper := strings.ToLower(body), strings.ToUpper(body)
	if body == lower || body == upper {
		return nil
	}
	hash := sha3.NewLegacyKeccak256()
	_, _ = hash.Write([]byte(lower))
	digest := hash.Sum(nil)
	for i, char := range body {
		if char >= '0' && char <= '9' {
			continue
		}
		nibble := digest[i/2]
		if i%2 == 0 {
			nibble >>= 4
		} else {
			nibble &= 0x0f
		}
		shouldUpper := nibble >= 8
		isUpper := char >= 'A' && char <= 'F'
		if shouldUpper != isUpper {
			return fmt.Errorf("mixed-case EVM address has invalid EIP-55 checksum")
		}
	}
	return nil
}
