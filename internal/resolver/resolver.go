package resolver

import (
	"context"
	"errors"
	"strings"
	"time"
)

type Policy struct{ AllowInsecure bool }

func StrictPolicy() Policy              { return Policy{} }
func InsecureDevelopmentPolicy() Policy { return Policy{AllowInsecure: true} }

type Resolver struct {
	transport   Transport
	validators  ValidatorRegistry
	defaultZone string
	clock       func() time.Time
}

func New(transport Transport, validators ValidatorRegistry, defaultZone string, clock func() time.Time) *Resolver {
	if clock == nil {
		clock = time.Now
	}
	return &Resolver{transport: transport, validators: validators, defaultZone: defaultZone, clock: clock}
}

func (r *Resolver) Resolve(ctx context.Context, name, chain string, policy Policy) (*Result, error) {
	queryName, err := CanonicalQueryName(name, r.defaultZone)
	if err != nil {
		return nil, &ResolveError{Code: CodeInvalidRecord, Name: name, Reason: err.Error(), Cause: err}
	}
	answer, err := r.transport.LookupTXT(ctx, queryName)
	if err != nil {
		var resolveErr *ResolveError
		if errors.As(err, &resolveErr) {
			return nil, resolveErr
		}
		return nil, &ResolveError{Code: CodeResolverUnavailable, Name: queryName, Reason: "validator lookup failed", Cause: err}
	}
	switch answer.Status {
	case DNSSECSecure:
	case DNSSECInsecure:
		if !policy.AllowInsecure {
			return nil, &ResolveError{Code: CodeDNSSECRequired, Name: queryName, Reason: "DNSSEC-secure answer required"}
		}
	case DNSSECBogus:
		return nil, &ResolveError{Code: CodeDNSSECBogus, Name: queryName, Reason: "validator reported a bogus DNSSEC chain"}
	default:
		return nil, &ResolveError{Code: CodeDNSSECRequired, Name: queryName, Reason: "DNSSEC status is indeterminate"}
	}
	records, err := ParseTXT(answer.Values)
	if err != nil {
		return nil, err
	}
	record, err := SelectChain(records, chain)
	if err != nil {
		return nil, err
	}
	if err := r.validators.Validate(chain, record.Address); err != nil {
		return nil, err
	}
	canonicalName := strings.TrimSuffix(strings.TrimPrefix(queryName, "_waddr."), ".")
	return &Result{
		Name: canonicalName, QueryName: queryName, Chain: chain, Address: record.Address,
		DNSSEC: answer.Status, Validator: r.transport.Validator(), TTL: answer.TTL,
		ResolvedAt: r.clock().UTC(),
	}, nil
}
