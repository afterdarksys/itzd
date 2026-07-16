package resolver

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"net"
	"testing"
	"time"

	"github.com/miekg/dns"
)

func startDoTServer(t *testing.T, handler dns.Handler) (string, *x509.CertPool) {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "validator.test"},
		DNSNames: []string{"validator.test"}, NotBefore: time.Now().Add(-time.Hour),
		NotAfter: time.Now().Add(time.Hour), KeyUsage: x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}, IsCA: true,
		BasicConstraintsValid: true,
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	cert := tls.Certificate{Certificate: [][]byte{der}, PrivateKey: key}
	pool := x509.NewCertPool()
	parsed, _ := x509.ParseCertificate(der)
	pool.AddCert(parsed)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	server := &dns.Server{
		Listener: tls.NewListener(listener, &tls.Config{Certificates: []tls.Certificate{cert}, MinVersion: tls.VersionTLS13}),
		Net:      "tcp", Handler: handler,
	}
	go func() { _ = server.ActivateAndServe() }()
	t.Cleanup(func() { _ = server.Shutdown() })
	return listener.Addr().String(), pool
}

func responseHandler(rcode int, authenticated, truncated bool) dns.Handler {
	return dns.HandlerFunc(func(w dns.ResponseWriter, request *dns.Msg) {
		response := new(dns.Msg)
		response.SetReply(request)
		response.Rcode = rcode
		response.AuthenticatedData = authenticated
		response.Truncated = truncated
		if rcode == dns.RcodeSuccess {
			response.Answer = append(response.Answer, &dns.TXT{
				Hdr: dns.RR_Header{Name: request.Question[0].Name, Rrtype: dns.TypeTXT, Class: dns.ClassINET, Ttl: 120},
				Txt: []string{"eip155:1:0x0000000000000000", "000000000000000000000001"},
			})
		}
		_ = w.WriteMsg(response)
	})
}

func TestDoTSecureTXTResponse(t *testing.T) {
	addr, roots := startDoTServer(t, responseHandler(dns.RcodeSuccess, true, false))
	transport, err := NewDoTTransport(DoTConfig{Endpoint: addr, ServerName: "validator.test", RootCAs: roots, Timeout: time.Second})
	if err != nil {
		t.Fatal(err)
	}
	answer, err := transport.LookupTXT(context.Background(), "_waddr.alice.example.")
	if err != nil {
		t.Fatal(err)
	}
	if answer.Status != DNSSECSecure || answer.TTL != 120 {
		t.Fatalf("unexpected answer %+v", answer)
	}
	if len(answer.Values) != 1 || answer.Values[0] != "eip155:1:0x0000000000000000000000000000000000000001" {
		t.Fatalf("unexpected TXT %#v", answer.Values)
	}
}

func TestDoTRejectsWrongTLSIdentity(t *testing.T) {
	addr, roots := startDoTServer(t, responseHandler(dns.RcodeSuccess, true, false))
	transport, _ := NewDoTTransport(DoTConfig{Endpoint: addr, ServerName: "wrong.test", RootCAs: roots, Timeout: time.Second})
	if _, err := transport.LookupTXT(context.Background(), "_waddr.alice.example."); err == nil {
		t.Fatal("expected TLS identity failure")
	}
}

func TestDoTTimeoutFailsClosed(t *testing.T) {
	handler := dns.HandlerFunc(func(w dns.ResponseWriter, request *dns.Msg) {
		time.Sleep(100 * time.Millisecond)
	})
	addr, roots := startDoTServer(t, handler)
	transport, _ := NewDoTTransport(DoTConfig{Endpoint: addr, ServerName: "validator.test", RootCAs: roots, Timeout: 10 * time.Millisecond})
	if _, err := transport.LookupTXT(context.Background(), "_waddr.alice.example."); err == nil {
		t.Fatal("expected timeout")
	}
}

func TestDoTMapsDNSSECAndResponseStates(t *testing.T) {
	tests := []struct {
		name          string
		rcode         int
		ad, truncated bool
		status        DNSSECStatus
		code          ErrorCode
	}{
		{"missing AD", dns.RcodeSuccess, false, false, DNSSECInsecure, ""},
		{"SERVFAIL", dns.RcodeServerFailure, false, false, DNSSECBogus, ""},
		{"NXDOMAIN", dns.RcodeNameError, true, false, "", CodeNameNotFound},
		{"truncated", dns.RcodeSuccess, true, true, "", CodeInvalidRecord},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addr, roots := startDoTServer(t, responseHandler(tt.rcode, tt.ad, tt.truncated))
			transport, _ := NewDoTTransport(DoTConfig{Endpoint: addr, ServerName: "validator.test", RootCAs: roots, Timeout: time.Second})
			answer, err := transport.LookupTXT(context.Background(), "_waddr.alice.example.")
			if tt.code != "" {
				resolveErr, ok := err.(*ResolveError)
				if !ok || resolveErr.Code != tt.code {
					t.Fatalf("got %v", err)
				}
				return
			}
			if err != nil || answer.Status != tt.status {
				t.Fatalf("answer=%+v err=%v", answer, err)
			}
		})
	}
}
