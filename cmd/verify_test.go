package cmd

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"itz.agency/itzd/internal/client"
	"itz.agency/itzd/internal/resolver"
)

type fakeParityClient struct {
	record *client.WalletRecord
	err    error
}

func (f fakeParityClient) ResolveNetwork(_, _ string) (*client.WalletRecord, error) {
	return f.record, f.err
}

func runVerify(t *testing.T, deps verifyDependencies, args ...string) (string, string, error) {
	t.Helper()
	cmd := newVerifyCommand(func() (verifyDependencies, error) { return deps, nil })
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return stdout.String(), stderr.String(), err
}

func TestVerifyReportsSecureDNSWithoutAPI(t *testing.T) {
	stdout, _, err := runVerify(t, verifyDependencies{resolver: &fakeResolver{result: secureResult()}}, "alice", "--chain", "eip155:1", "--json")
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{`"dnssec": "secure"`, `"address_valid": true`, `"validator"`} {
		if !strings.Contains(stdout, expected) {
			t.Fatalf("missing %s in %s", expected, stdout)
		}
	}
}

func TestVerifyAPIMismatchIsDiagnosticFailure(t *testing.T) {
	deps := verifyDependencies{resolver: &fakeResolver{result: secureResult()}, api: fakeParityClient{record: &client.WalletRecord{WalletAddress: "0x0000000000000000000000000000000000000002"}}}
	_, _, err := runVerify(t, deps, "alice", "--chain", "eip155:1", "--check-api")
	var resolveErr *resolver.ResolveError
	if !errors.As(err, &resolveErr) || resolveErr.Code != resolver.CodeAPIDNSMismatch {
		t.Fatalf("got %v", err)
	}
}

func TestVerifyAPIUnreachableDoesNotInvalidateSecureDNS(t *testing.T) {
	deps := verifyDependencies{resolver: &fakeResolver{result: secureResult()}, api: fakeParityClient{err: errors.New("offline")}}
	stdout, _, err := runVerify(t, deps, "alice", "--chain", "eip155:1", "--check-api", "--json")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdout, `"api_reachable": false`) {
		t.Fatalf("unexpected report %s", stdout)
	}
}

func TestVerifyDNSFailureReturnsNoReport(t *testing.T) {
	deps := verifyDependencies{resolver: &fakeResolver{err: &resolver.ResolveError{Code: resolver.CodeDNSSECRequired}}}
	stdout, _, err := runVerify(t, deps, "alice", "--chain", "eip155:1")
	if err == nil || stdout != "" {
		t.Fatalf("stdout=%q err=%v", stdout, err)
	}
}

var _ resolveService = (*fakeResolver)(nil)
