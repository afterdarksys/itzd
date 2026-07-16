package resolver

import "testing"

func TestErrorCodeIsStable(t *testing.T) {
	err := &ResolveError{Code: CodeDNSSECRequired, Name: "alice.example"}
	if got := err.Error(); got == "" {
		t.Fatal("expected a human-readable error")
	}
	if err.Code != "DNSSEC_REQUIRED" {
		t.Fatalf("unexpected code %q", err.Code)
	}
}

func TestDNSSECStatusValuesAreStable(t *testing.T) {
	want := map[DNSSECStatus]string{
		DNSSECSecure: "secure", DNSSECInsecure: "insecure",
		DNSSECBogus: "bogus", DNSSECIndeterminate: "indeterminate",
	}
	for status, value := range want {
		if string(status) != value {
			t.Fatalf("%q != %q", status, value)
		}
	}
}
