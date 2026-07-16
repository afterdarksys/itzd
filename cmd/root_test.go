package cmd

import (
	"fmt"
	"testing"

	"itz.agency/itzd/internal/resolver"
)

func TestStableResolverExitCodes(t *testing.T) {
	tests := map[resolver.ErrorCode]int{
		resolver.CodeNameNotFound: 2, resolver.CodeChainNotFound: 3,
		resolver.CodeDNSSECRequired: 4, resolver.CodeDNSSECBogus: 5,
		resolver.CodeInvalidRecord: 6, resolver.CodeInvalidAddress: 7,
		resolver.CodeConflictingRecords: 8, resolver.CodeResolverUnavailable: 9,
		resolver.CodeAPIDNSMismatch: 10,
	}
	for code, want := range tests {
		if got := exitCode(&resolver.ResolveError{Code: code}); got != want {
			t.Fatalf("%s=%d want %d", code, got, want)
		}
		if got := exitCode(fmt.Errorf("wrapped: %w", &resolver.ResolveError{Code: code})); got != want {
			t.Fatalf("wrapped %s=%d", code, got)
		}
	}
}
