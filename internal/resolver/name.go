package resolver

import (
	"fmt"
	"net"
	"strings"

	"github.com/miekg/dns"
)

func CanonicalQueryName(input, defaultZone string) (string, error) {
	name := strings.ToLower(strings.TrimSpace(input))
	name = strings.TrimSuffix(name, ".")
	if name == "" || name == "_waddr" {
		return "", fmt.Errorf("name is empty")
	}
	for _, r := range name {
		if r < 0x21 || r > 0x7e {
			return "", fmt.Errorf("name must contain printable ASCII only")
		}
	}
	if net.ParseIP(name) != nil {
		return "", fmt.Errorf("IP literals are not wallet names")
	}
	if strings.HasPrefix(name, "_waddr.") {
		name = strings.TrimPrefix(name, "_waddr.")
	}
	if !strings.Contains(name, ".") {
		zone := strings.Trim(strings.ToLower(strings.TrimSpace(defaultZone)), ".")
		if zone == "" {
			return "", fmt.Errorf("default zone is empty")
		}
		name += "." + zone
	}
	for _, label := range strings.Split(name, ".") {
		if label == "" || len(label) > 63 {
			return "", fmt.Errorf("invalid DNS label")
		}
		for i, r := range label {
			if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || (r == '-' && i > 0 && i < len(label)-1) {
				continue
			}
			return "", fmt.Errorf("invalid character in DNS label")
		}
	}
	query := dns.Fqdn("_waddr." + name)
	if _, ok := dns.IsDomainName(query); !ok {
		return "", fmt.Errorf("invalid DNS name")
	}
	return query, nil
}
