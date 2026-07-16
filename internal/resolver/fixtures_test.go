package resolver

import (
	"os"
	"strings"
	"testing"
)

func TestSecureZoneFixtureIsSignedAndClearlyTestOnly(t *testing.T) {
	signed, err := os.ReadFile("../../testdata/secure-zone/zone.signed")
	if err != nil {
		t.Fatal(err)
	}
	contents := string(signed)
	if !strings.Contains(contents, "RRSIG") || !strings.Contains(contents, "DNSKEY") {
		t.Fatal("fixture is not DNSSEC signed")
	}
	readme, err := os.ReadFile("../../testdata/secure-zone/README.md")
	if err != nil || !strings.Contains(strings.ToLower(string(readme)), "test-only") {
		t.Fatal("fixture keys lack test-only warning")
	}
}
