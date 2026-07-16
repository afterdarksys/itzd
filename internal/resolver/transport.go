package resolver

import "context"

type TXTAnswer struct {
	Values []string
	TTL    uint32
	Status DNSSECStatus
}

type Transport interface {
	LookupTXT(ctx context.Context, fqdn string) (TXTAnswer, error)
	Validator() string
}
