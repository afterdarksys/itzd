package resolver

import (
	"regexp"
	"strings"
)

var (
	namespacePattern = regexp.MustCompile(`^[-a-z0-9]{3,8}$`)
	referencePattern = regexp.MustCompile(`^[-_a-zA-Z0-9]{1,32}$`)
)

type Record struct {
	Chain   string
	Address string
}

func ParseTXT(values []string) ([]Record, error) {
	seen := make(map[string]struct{})
	records := make([]Record, 0, len(values))
	for _, value := range values {
		if value == "" || len(value) > 1024 || strings.TrimSpace(value) != value {
			return nil, &ResolveError{Code: CodeInvalidRecord, Reason: "invalid TXT length or whitespace"}
		}
		parts := strings.SplitN(value, ":", 3)
		if len(parts) != 3 || !namespacePattern.MatchString(parts[0]) ||
			!referencePattern.MatchString(parts[1]) || parts[2] == "" {
			return nil, &ResolveError{Code: CodeInvalidRecord, Reason: "malformed wallet TXT record"}
		}
		record := Record{Chain: parts[0] + ":" + parts[1], Address: parts[2]}
		key := record.Chain + "\x00" + record.Address
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		records = append(records, record)
	}
	if len(records) == 0 {
		return nil, &ResolveError{Code: CodeInvalidRecord, Reason: "no wallet TXT records"}
	}
	return records, nil
}

func SelectChain(records []Record, chain string) (Record, error) {
	var selected *Record
	for i := range records {
		if records[i].Chain != chain {
			continue
		}
		if selected != nil && selected.Address != records[i].Address {
			return Record{}, &ResolveError{Code: CodeConflictingRecords, Chain: chain, Reason: "multiple addresses for chain"}
		}
		copy := records[i]
		selected = &copy
	}
	if selected == nil {
		return Record{}, &ResolveError{Code: CodeChainNotFound, Chain: chain, Reason: "requested chain is absent"}
	}
	return *selected, nil
}
