package resolver

import "time"

type DNSSECStatus string

const (
	DNSSECSecure        DNSSECStatus = "secure"
	DNSSECInsecure      DNSSECStatus = "insecure"
	DNSSECBogus         DNSSECStatus = "bogus"
	DNSSECIndeterminate DNSSECStatus = "indeterminate"
)

type Result struct {
	Name       string       `json:"name"`
	QueryName  string       `json:"query_name"`
	Chain      string       `json:"chain"`
	Address    string       `json:"address"`
	DNSSEC     DNSSECStatus `json:"dnssec"`
	Validator  string       `json:"validator"`
	TTL        uint32       `json:"ttl"`
	ResolvedAt time.Time    `json:"resolved_at"`
}
