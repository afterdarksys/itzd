package cmd

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"itz.agency/itzd/internal/resolver"
)

type fakeResolver struct {
	result      *resolver.Result
	err         error
	name, chain string
	policy      resolver.Policy
}

func (f *fakeResolver) Resolve(_ context.Context, name, chain string, policy resolver.Policy) (*resolver.Result, error) {
	f.name, f.chain, f.policy = name, chain, policy
	return f.result, f.err
}

func runResolve(t *testing.T, service resolveService, args ...string) (string, string, error) {
	t.Helper()
	cmd := newResolveCommand(func() (resolveService, error) { return service, nil })
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return stdout.String(), stderr.String(), err
}

func secureResult() *resolver.Result {
	return &resolver.Result{Name: "alice.itz.agency", QueryName: "_waddr.alice.itz.agency.", Chain: "eip155:1", Address: "0x0000000000000000000000000000000000000001", DNSSEC: resolver.DNSSECSecure, Validator: "validator.test", TTL: 300, ResolvedAt: time.Unix(100, 0).UTC()}
}

func TestResolveHumanAndJSONOutput(t *testing.T) {
	for _, jsonOutput := range []bool{false, true} {
		fake := &fakeResolver{result: secureResult()}
		args := []string{"alice", "--chain", "eip155:1"}
		if jsonOutput {
			args = append(args, "--json")
		}
		stdout, _, err := runResolve(t, fake, args...)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(stdout, secureResult().Address) {
			t.Fatalf("missing address: %s", stdout)
		}
		if jsonOutput && !strings.Contains(stdout, `"dnssec": "secure"`) {
			t.Fatalf("missing JSON provenance: %s", stdout)
		}
	}
}

func TestResolveStrictFailurePrintsNoAddress(t *testing.T) {
	fake := &fakeResolver{err: &resolver.ResolveError{Code: resolver.CodeDNSSECRequired, Reason: "secure required"}}
	stdout, _, err := runResolve(t, fake, "alice", "--chain", "eip155:1")
	if err == nil {
		t.Fatal("expected error")
	}
	if stdout != "" {
		t.Fatalf("failure wrote stdout: %q", stdout)
	}
}

func TestResolveAllowInsecureWarnsAndSetsPolicy(t *testing.T) {
	result := secureResult()
	result.DNSSEC = resolver.DNSSECInsecure
	fake := &fakeResolver{result: result}
	stdout, stderr, err := runResolve(t, fake, "alice.example", "--chain", "eip155:1", "--allow-insecure", "--json")
	if err != nil {
		t.Fatal(err)
	}
	if !fake.policy.AllowInsecure || !strings.Contains(stderr, "WARNING") {
		t.Fatalf("policy=%+v stderr=%q", fake.policy, stderr)
	}
	if !strings.Contains(stdout, `"dnssec": "insecure"`) {
		t.Fatalf("unexpected JSON %s", stdout)
	}
	if fake.name != "alice.example" {
		t.Fatalf("custom domain changed to %q", fake.name)
	}
}

func TestResolveDeprecatedNetworkAlias(t *testing.T) {
	fake := &fakeResolver{result: secureResult()}
	_, stderr, err := runResolve(t, fake, "alice", "--network", "e")
	if err != nil {
		t.Fatal(err)
	}
	if fake.chain != "eip155:1" || !strings.Contains(stderr, "deprecated") {
		t.Fatalf("chain=%q stderr=%q", fake.chain, stderr)
	}
}

func TestResolveRequiresChainAndPropagatesBuilderFailure(t *testing.T) {
	if _, _, err := runResolve(t, &fakeResolver{}, "alice"); err == nil {
		t.Fatal("expected chain error")
	}
	cmd := newResolveCommand(func() (resolveService, error) { return nil, errors.New("config failed") })
	cmd.SetArgs([]string{"alice", "--chain", "eip155:1"})
	if err := cmd.Execute(); err == nil {
		t.Fatal("expected builder error")
	}
}
