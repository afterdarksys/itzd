package resolver

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakeTransport struct {
	answer TXTAnswer
	err    error
}

func (f fakeTransport) LookupTXT(context.Context, string) (TXTAnswer, error) { return f.answer, f.err }
func (f fakeTransport) Validator() string                                    { return "validator.test@127.0.0.1:853" }

func newResolverWithAnswer(answer TXTAnswer) *Resolver {
	return New(fakeTransport{answer: answer}, DefaultValidatorRegistry(), "itz.agency", func() time.Time { return time.Unix(100, 0).UTC() })
}

func TestStrictPolicyReturnsNoAddressForInsecureDNS(t *testing.T) {
	for _, status := range []DNSSECStatus{DNSSECInsecure, DNSSECIndeterminate, DNSSECBogus} {
		r := newResolverWithAnswer(TXTAnswer{Values: []string{"eip155:1:0x0000000000000000000000000000000000000001"}, Status: status})
		result, err := r.Resolve(context.Background(), "alice", "eip155:1", StrictPolicy())
		if err == nil || result != nil {
			t.Fatalf("status %s returned result %+v err=%v", status, result, err)
		}
	}
}

func TestInsecureDevelopmentPolicyIsExplicit(t *testing.T) {
	r := newResolverWithAnswer(TXTAnswer{Values: []string{"eip155:1:0x0000000000000000000000000000000000000001"}, Status: DNSSECInsecure, TTL: 60})
	result, err := r.Resolve(context.Background(), "alice", "eip155:1", InsecureDevelopmentPolicy())
	if err != nil {
		t.Fatal(err)
	}
	if result.DNSSEC != DNSSECInsecure || result.Address == "" {
		t.Fatalf("unexpected result %+v", result)
	}
}

func TestPolicyFailuresNeverReturnPartialResult(t *testing.T) {
	tests := []struct {
		name      string
		transport fakeTransport
		chain     string
		code      ErrorCode
	}{
		{"timeout", fakeTransport{err: errors.New("timeout")}, "eip155:1", CodeResolverUnavailable},
		{"conflict", fakeTransport{answer: TXTAnswer{Values: []string{"eip155:1:0x0000000000000000000000000000000000000001", "eip155:1:0x0000000000000000000000000000000000000002"}, Status: DNSSECSecure}}, "eip155:1", CodeConflictingRecords},
		{"invalid address", fakeTransport{answer: TXTAnswer{Values: []string{"eip155:1:0x1234"}, Status: DNSSECSecure}}, "eip155:1", CodeInvalidAddress},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := New(tt.transport, DefaultValidatorRegistry(), "itz.agency", time.Now)
			result, err := r.Resolve(context.Background(), "alice", tt.chain, StrictPolicy())
			if result != nil {
				t.Fatalf("failure returned %+v", result)
			}
			var resolveErr *ResolveError
			if !errors.As(err, &resolveErr) || resolveErr.Code != tt.code {
				t.Fatalf("got %v", err)
			}
		})
	}
}

func TestSecurePolicyBuildsProvenance(t *testing.T) {
	r := newResolverWithAnswer(TXTAnswer{Values: []string{"eip155:1:0x0000000000000000000000000000000000000001"}, Status: DNSSECSecure, TTL: 300})
	result, err := r.Resolve(context.Background(), "alice", "eip155:1", StrictPolicy())
	if err != nil {
		t.Fatal(err)
	}
	if result.QueryName != "_waddr.alice.itz.agency." || result.Validator == "" || result.TTL != 300 {
		t.Fatalf("unexpected result %+v", result)
	}
}
