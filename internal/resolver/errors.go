package resolver

import "fmt"

type ErrorCode string

const (
	CodeNameNotFound        ErrorCode = "NAME_NOT_FOUND"
	CodeChainNotFound       ErrorCode = "CHAIN_NOT_FOUND"
	CodeDNSSECRequired      ErrorCode = "DNSSEC_REQUIRED"
	CodeDNSSECBogus         ErrorCode = "DNSSEC_BOGUS"
	CodeInvalidRecord       ErrorCode = "INVALID_RECORD"
	CodeInvalidAddress      ErrorCode = "INVALID_ADDRESS"
	CodeConflictingRecords  ErrorCode = "CONFLICTING_RECORDS"
	CodeResolverUnavailable ErrorCode = "RESOLVER_UNAVAILABLE"
	CodeAPIDNSMismatch      ErrorCode = "API_DNS_MISMATCH"
)

type ResolveError struct {
	Code   ErrorCode `json:"code"`
	Name   string    `json:"name,omitempty"`
	Chain  string    `json:"chain,omitempty"`
	Reason string    `json:"reason,omitempty"`
	Cause  error     `json:"-"`
}

func (e *ResolveError) Error() string {
	detail := e.Reason
	if detail == "" {
		detail = "resolution failed"
	}
	if e.Name != "" {
		return fmt.Sprintf("%s for %s: %s", e.Code, e.Name, detail)
	}
	return fmt.Sprintf("%s: %s", e.Code, detail)
}

func (e *ResolveError) Unwrap() error { return e.Cause }
