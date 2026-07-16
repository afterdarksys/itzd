package resolver

import (
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"
)

type conformanceCase struct {
	Name            string     `json:"name"`
	TXT             [][]string `json:"txt"`
	RequestedChain  string     `json:"requested_chain"`
	ExpectedAddress string     `json:"expected_address"`
	ExpectedError   ErrorCode  `json:"expected_error"`
}

func TestRecordConformance(t *testing.T) {
	data, err := os.ReadFile("testdata/records.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []conformanceCase
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			values := make([]string, len(tc.TXT))
			for i, chunks := range tc.TXT {
				values[i] = strings.Join(chunks, "")
			}
			records, err := ParseTXT(values)
			if err == nil {
				_, err = SelectChain(records, tc.RequestedChain)
			}
			if tc.ExpectedError != "" {
				var resolveErr *ResolveError
				if !errors.As(err, &resolveErr) || resolveErr.Code != tc.ExpectedError {
					t.Fatalf("got %v want %s", err, tc.ExpectedError)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			record, _ := SelectChain(records, tc.RequestedChain)
			if record.Address != tc.ExpectedAddress {
				t.Fatalf("got %q", record.Address)
			}
		})
	}
}

func TestParseTXTRejectsOversizedValue(t *testing.T) {
	_, err := ParseTXT([]string{"eip155:1:" + strings.Repeat("a", 1025)})
	if err == nil {
		t.Fatal("expected oversized record error")
	}
}
