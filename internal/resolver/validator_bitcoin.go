package resolver

import (
	"fmt"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
)

type BitcoinMainnetValidator struct{}

func (BitcoinMainnetValidator) Validate(address string) error {
	decoded, err := btcutil.DecodeAddress(address, &chaincfg.MainNetParams)
	if err != nil {
		return fmt.Errorf("invalid Bitcoin address")
	}
	if !decoded.IsForNet(&chaincfg.MainNetParams) {
		return fmt.Errorf("Bitcoin address is not mainnet")
	}
	return nil
}
