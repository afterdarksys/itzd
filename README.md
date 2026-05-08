# itzd — DNS-Native Wallet Identity for Every Chain

**Send crypto to `ryan.itz.agency` on Ethereum, Bitcoin, Solana, or any of 24+ chains.**

No gas fees. No blockchain. Standard DNS. Works in 30 seconds.

```
$ itzd register ryan e 0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb
✓ Registered: ryan.itz.agency
  Network  : e (eip155:1)
  Wallet   : 0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb
  DNS name : _waddr.ryan.itz.agency
  DNS sync : ✓ live

$ itzd resolve ryan
ryan.itz.agency  (3 wallets)
────────────────────────────────────────────────────────────
  e      0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb  ✓ verified
         DNS: _waddr.ryan.itz.agency
         TXT: eip155:1:0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb
  b      bc1qxy2kgdygjrsqtzq2n0yrf2493p83kkfjhx0wlh
         DNS: _waddr.ryan.itz.agency
         TXT: bip122:btc:bc1qxy2kgdygjrsqtzq2n0yrf2493p83kkfjhx0wlh
  s      DRpbCBMxVnDK7maPM5tGv6MvB3v1sRMC86PZ8okm21hy
         DNS: _waddr.ryan.itz.agency
         TXT: solana:sol:DRpbCBMxVnDK7maPM5tGv6MvB3v1sRMC86PZ8okm21hy

$ dig TXT _waddr.ryan.itz.agency +short
"eip155:1:0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb"
"bip122:btc:bc1qxy2kgdygjrsqtzq2n0yrf2493p83kkfjhx0wlh"
"solana:sol:DRpbCBMxVnDK7maPM5tGv6MvB3v1sRMC86PZ8okm21hy"
```

---

## What Is This?

**itzd** is the client for [itz.agency](https://itz.agency) — a production DNS nameserver infrastructure implementing
[draft-chins-dnsop-web3-wallet-mapping-03](https://datatracker.ietf.org/doc/html/draft-chins-dnsop-web3-wallet-mapping-03),
the active IETF standard for DNS-to-wallet address mapping.

The complete stack:

| Component | Role |
|-----------|------|
| **itz.agency** | Registration platform — claim your name, link your wallets |
| **itzd** | CLI client — register, resolve, manage from the terminal |
| **dnsscienced** | DNS nameserver — serves `_waddr` TXT records authoritatively |
| **Resolution API** | `GET /resolve/{name}` — drop-in HTTP resolution for any app |

---

## Quick Start

### Install

```bash
# macOS (Apple Silicon)
curl -Lo itzd https://github.com/afterdark/itzd/releases/latest/download/itzd-darwin-arm64
chmod +x itzd && sudo mv itzd /usr/local/bin/

# macOS (Intel)
curl -Lo itzd https://github.com/afterdark/itzd/releases/latest/download/itzd-darwin-amd64
chmod +x itzd && sudo mv itzd /usr/local/bin/

# Linux (amd64)
curl -Lo itzd https://github.com/afterdark/itzd/releases/latest/download/itzd-linux-amd64
chmod +x itzd && sudo mv itzd /usr/local/bin/
```

### Register a Name

```bash
itzd login                                               # authenticate with your wallet
itzd register yourname e 0xYOUR_ETH_WALLET              # Ethereum
itzd register yourname b YOUR_BTC_ADDRESS               # Bitcoin
itzd register yourname s YOUR_SOLANA_ADDRESS            # Solana
```

### Resolve

```bash
itzd resolve yourname           # all chains
itzd resolve yourname -n e      # Ethereum only
itzd whoami 0xYOUR_ADDRESS      # reverse: wallet → names
```

### Check DNS directly (no itzd required)

```bash
dig TXT _waddr.yourname.itz.agency +short
```

---

## REST API

The resolution API is **public, unauthenticated, and CDN-cacheable**.

```bash
# Resolve all wallets for a username
GET https://api.itz.agency/resolve/{username}

# Resolve a specific chain
GET https://api.itz.agency/resolve/{username}/{network}

# Reverse lookup: wallet → usernames
GET https://api.itz.agency/resolve/wallet/{address}
```

### Example responses

```bash
$ curl https://api.itz.agency/resolve/ryan/e
{
  "username": "ryan",
  "network": "e",
  "caip2": "eip155:1",
  "wallet_address": "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
  "txt_record": "eip155:1:0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
  "dns_name": "_waddr.ryan.itz.agency",
  "verified": true,
  "dns_synced": true
}

$ curl https://api.itz.agency/resolve/wallet/0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb
{
  "wallet_address": "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
  "usernames": [{"username": "ryan", "network": "e", ...}],
  "count": 1
}
```

---

## Integration (15 minutes or less)

### JavaScript / TypeScript

```typescript
async function resolveWallet(username: string, network = 'e') {
  const res = await fetch(`https://api.itz.agency/resolve/${username}/${network}`)
  if (!res.ok) return null
  const data = await res.json()
  return data.wallet_address
}

// In your send flow:
const wallet = await resolveWallet('ryan', 'e')
// → "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb"
```

### Python

```python
import httpx

def resolve(username: str, network: str = "e") -> str | None:
    r = httpx.get(f"https://api.itz.agency/resolve/{username}/{network}")
    return r.json()["wallet_address"] if r.status_code == 200 else None
```

### Go

```go
resp, err := http.Get("https://api.itz.agency/resolve/" + username + "/e")
// parse JSON, read wallet_address
```

### DNS (no HTTP, any language)

```python
import dns.resolver
answers = dns.resolver.resolve(f"_waddr.{username}.itz.agency", "TXT")
# parse "eip155:1:0xABC..." format per IETF draft
```

---

## Supported Chains (24+)

| Code | Chain | CAIP-2 |
|------|-------|--------|
| `e` | Ethereum | `eip155:1` |
| `b` | Bitcoin | `bip122:btc` |
| `s` | Solana | `solana:sol` |
| `poly` | Polygon | `eip155:137` |
| `arb` | Arbitrum | `eip155:42161` |
| `op` | Optimism | `eip155:10` |
| `cb` | Base | `eip155:8453` |
| `bnb` | BNB Chain | `eip155:56` |
| `atom` | Cosmos Hub | `cosmos:cosmoshub-4` |
| `dot` | Polkadot | `polkadot:91b171bb` |
| `ada` | Cardano | `cardano:mainnet` |
| `hbar` | Hedera | `hedera:mainnet` |
| ... | 12 more | — |

Run `itzd networks` for the full list.

---

## Standards

- **[draft-chins-dnsop-web3-wallet-mapping-03](https://datatracker.ietf.org/doc/html/draft-chins-dnsop-web3-wallet-mapping-03)** — IETF standards-track draft
- **[CAIP-2](https://github.com/ChainAgnostic/CAIPs/blob/master/CAIPs/caip-2.md)** — Blockchain ID specification
- **[SLIP-0044](https://github.com/satoshilabs/slips/blob/master/slip-0044.md)** — Registered coin types
- **[RFC 8945](https://www.rfc-editor.org/rfc/rfc8945.html)** — DNS TSIG

---

## Pricing

| Plan | Names | Chains | API Lookups/day | Price |
|------|-------|--------|-----------------|-------|
| **Free** | 1 | 3 | 100 | $0 |
| **Pro** | 10 | All 24+ | 10,000 | $9/mo |
| **Business** | Unlimited | All | 100,000 | $49/mo |
| **Enterprise** | Unlimited | All | Unlimited + SLA | Contact |

[Get started at itz.agency](https://itz.agency)

---

## For Wallets & Exchanges

We support white-label integration for:

- **Wallet apps** — let users type `name.itz.agency` in the send field
- **Exchanges** — resolve withdrawal addresses from human-readable names
- **Payment processors** — multi-chain address resolution in one call
- **Enterprise** — company namespace (`employee.company.itz.agency`) for crypto payroll

API access, SLAs, and volume pricing available. Contact: `api@itz.agency`

---

## Architecture

```
User registers "ryan.itz.agency" [Ethereum]
    │
    ▼
itz.agency API (FastAPI, PostgreSQL)
    │  stores wallet mapping
    │  writes _waddr TXT record via gRPC
    ▼
dnsscienced nameserver (ns1.idoms.net, ns2.idoms.net)
    │  serves TXT record authoritative
    │  implements draft-chins-dnsop-web3-wallet-mapping-03
    ▼
Any DNS resolver in the world
    │
    ▼
Any wallet, exchange, or app resolves the name
```

---

*itz.agency is a production service running on custom DNS infrastructure implementing an active IETF standards-track draft. This is not another blockchain naming system — it's DNS, the internet's existing address book.*
