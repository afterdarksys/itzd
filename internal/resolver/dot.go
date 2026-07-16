package resolver

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"time"

	"github.com/miekg/dns"
)

type DoTConfig struct {
	Endpoint   string
	ServerName string
	RootCAs    *x509.CertPool
	Timeout    time.Duration
}

type DoTTransport struct {
	endpoint   string
	serverName string
	client     *dns.Client
}

func NewDoTTransport(cfg DoTConfig) (*DoTTransport, error) {
	if _, _, err := net.SplitHostPort(cfg.Endpoint); err != nil {
		return nil, fmt.Errorf("invalid validator endpoint: %w", err)
	}
	if cfg.ServerName == "" {
		return nil, fmt.Errorf("validator TLS server name is required")
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 5 * time.Second
	}
	return &DoTTransport{
		endpoint: cfg.Endpoint, serverName: cfg.ServerName,
		client: &dns.Client{
			Net: "tcp-tls", Timeout: cfg.Timeout,
			TLSConfig: &tls.Config{ServerName: cfg.ServerName, RootCAs: cfg.RootCAs, MinVersion: tls.VersionTLS13},
		},
	}, nil
}

func DefaultDoTTransport() (*DoTTransport, error) {
	return NewDoTTransport(DoTConfig{
		Endpoint: "1.1.1.1:853", ServerName: "cloudflare-dns.com", Timeout: 5 * time.Second,
	})
}

func (t *DoTTransport) Validator() string { return t.serverName + "@" + t.endpoint }

func (t *DoTTransport) LookupTXT(ctx context.Context, fqdn string) (TXTAnswer, error) {
	request := new(dns.Msg)
	request.SetQuestion(fqdn, dns.TypeTXT)
	request.SetEdns0(1232, true)
	response, _, err := t.client.ExchangeContext(ctx, request, t.endpoint)
	if err != nil {
		return TXTAnswer{}, &ResolveError{Code: CodeResolverUnavailable, Name: fqdn, Reason: "authenticated validator unavailable", Cause: err}
	}
	if response.Truncated {
		return TXTAnswer{}, &ResolveError{Code: CodeInvalidRecord, Name: fqdn, Reason: "truncated DNS-over-TLS response"}
	}
	if response.Rcode == dns.RcodeNameError {
		return TXTAnswer{}, &ResolveError{Code: CodeNameNotFound, Name: fqdn, Reason: "name does not exist"}
	}
	if response.Rcode == dns.RcodeServerFailure {
		return TXTAnswer{Status: DNSSECBogus}, nil
	}
	if response.Rcode != dns.RcodeSuccess {
		return TXTAnswer{}, &ResolveError{Code: CodeResolverUnavailable, Name: fqdn, Reason: "validator returned DNS error"}
	}

	answer := TXTAnswer{Status: DNSSECInsecure}
	if response.AuthenticatedData {
		answer.Status = DNSSECSecure
	}
	var ttl uint32
	for _, rr := range response.Answer {
		txt, ok := rr.(*dns.TXT)
		if !ok {
			continue
		}
		answer.Values = append(answer.Values, joinTXTChunks(txt.Txt))
		if ttl == 0 || txt.Hdr.Ttl < ttl {
			ttl = txt.Hdr.Ttl
		}
	}
	if len(answer.Values) == 0 {
		return TXTAnswer{}, &ResolveError{Code: CodeNameNotFound, Name: fqdn, Reason: "wallet TXT record does not exist"}
	}
	answer.TTL = ttl
	return answer, nil
}

func joinTXTChunks(chunks []string) string {
	var result string
	for _, chunk := range chunks {
		result += chunk
	}
	return result
}
