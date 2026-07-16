# itzd

`itzd` is a fail-closed cryptocurrency-address resolver for `_waddr` DNS TXT
records. It queries an explicitly configured validating resolver over
certificate-verified DNS-over-TLS, requires a DNSSEC-secure response, selects an
exact CAIP-2 chain, and validates the returned address before printing it.

```text
alice.itz.agency
  -> _waddr.alice.itz.agency. TXT
  -> eip155:1:0x0000000000000000000000000000000000000001
```

## Trust model

The current release trusts an authenticated external validating resolver. It
does not perform native iterative DNSSEC validation locally. The validator
endpoint and TLS server name are included in every result. The ITZ HTTP API is
never an automatic fallback and is used only by `verify --check-api`.

No address is returned for `insecure`, `bogus`, or `indeterminate` DNSSEC state.
`--allow-insecure` exists only for development, emits a warning on stderr, and
marks JSON output insecure. It never accepts a bogus chain.

## Build and test

Go 1.26 or newer is required.

```bash
go build ./...
go test -race ./...
go vet ./...
```

## Resolve

```bash
itzd resolve alice --chain eip155:1
itzd resolve alice.itz.agency --chain eip155:1 --json
itzd resolve alice.example.com --chain eip155:1
```

A single-label name uses `itz.agency` as its default zone. Multi-label names are
custom domains and are never silently rewritten. The deprecated aliases
`--network e`, `--network b`, and `--network s` remain temporarily available
and print a warning.

## Diagnose

```bash
itzd verify alice --chain eip155:1 --json
itzd verify alice --chain eip155:1 --json --check-api
```

The report includes query name, chain, address, DNSSEC status, validator, TTL,
resolution time, address validation, API reachability, and API/DNS parity. API
downtime does not invalidate secure DNS; a mismatch exits nonzero.

## Configuration

Configuration follows the XDG user configuration directory and is saved with
mode `0600`. Environment values override disk:

```text
ITZ_API
ITZ_TOKEN
ITZ_VALIDATOR_ENDPOINT       default: 1.1.1.1:853
ITZ_VALIDATOR_SERVER_NAME    default: cloudflare-dns.com
ITZ_DEFAULT_ZONE             default: itz.agency
```

For noninteractive authentication, obtain a token through an audited client and
set `ITZ_TOKEN`. The CLI does not fabricate or locally custody wallet keys.

## Audited address validation

The initial registry validates Ethereum mainnet address syntax and EIP-55,
Solana 32-byte Base58 public keys, and Bitcoin mainnet addresses. Other exact
CAIP-2 identifiers fail closed until a reviewed validator is added.

## Stable resolver exit codes

| Exit | Error code |
|---:|---|
| 2 | `NAME_NOT_FOUND` |
| 3 | `CHAIN_NOT_FOUND` |
| 4 | `DNSSEC_REQUIRED` |
| 5 | `DNSSEC_BOGUS` |
| 6 | `INVALID_RECORD` |
| 7 | `INVALID_ADDRESS` |
| 8 | `CONFLICTING_RECORDS` |
| 9 | `RESOLVER_UNAVAILABLE` |
| 10 | `API_DNS_MISMATCH` |

Versioned parser fixtures live under `internal/resolver/testdata`. Production
examples must not be described as secure until the hosted parent DS and canary
verification gates pass.
