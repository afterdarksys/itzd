package resolver

import "testing"

func TestCanonicalQueryName(t *testing.T) {
	tests := []struct{ input, want string }{
		{"alice", "_waddr.alice.itz.agency."},
		{"alice.itz.agency", "_waddr.alice.itz.agency."},
		{"ALICE.ITZ.AGENCY.", "_waddr.alice.itz.agency."},
		{"alice.example.com.", "_waddr.alice.example.com."},
		{"_waddr.alice.example.com", "_waddr.alice.example.com."},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := CanonicalQueryName(tt.input, "itz.agency")
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Fatalf("got %q want %q", got, tt.want)
			}
		})
	}
}

func TestCanonicalQueryNameRejectsAmbiguousInput(t *testing.T) {
	long := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa.example"
	for _, input := range []string{"", ".", "alice..example", long, "127.0.0.1", "alice\n.example", "álîce.example", "_waddr"} {
		t.Run(input, func(t *testing.T) {
			if got, err := CanonicalQueryName(input, "itz.agency"); err == nil {
				t.Fatalf("expected error, got %q", got)
			}
		})
	}
}
