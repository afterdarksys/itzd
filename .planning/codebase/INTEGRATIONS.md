# External Integrations

**Analysis Date:** 2026-05-08

## APIs & External Services

**itz.agency API:**
- itz.agency Resolution API - Primary integration for DNS-to-wallet mapping, registration, and verification
  - SDK/Client: Custom HTTP client (`internal/client/client.go`)
  - Auth: Bearer token (stored in `~/.itzd/config.json`)
  - Endpoints: `/health`, `/auth/challenge`, `/auth/verify/ethereum`, `/resolve/{username}`, `/resolve/{username}/{network}`, `/resolve/wallet/{address}`, `/username/register`, `/username/check`, `/username/my-usernames`

**DNS Resolution:**
- Standard DNS infrastructure (read-only, via system DNS resolver)
  - Used to verify DNS TXT records at `_waddr.{username}.itz.agency`
  - Implements draft-chins-dnsop-web3-wallet-mapping-03 standard
  - No external SDK required - uses standard library `net` package DNS resolution

## Data Storage

**Databases:**
- None (CLI-only) - State is ephemeral and stored in API

**Local Configuration:**
- File storage: `~/.itzd/config.json`
  - Format: JSON
  - Contains: API endpoint, authentication token, wallet address, preferred network

**File Storage:**
- No external file storage integration

**Caching:**
- None implemented - all resolution queries hit the API

## Authentication & Identity

**Auth Provider:**
- Wallet-based authentication via itz.agency
  - Implementation: Challenge-response signing
  - Supported chains: Ethereum (EIP-191 signature scheme)
  - Flow:
    1. `POST /auth/challenge` - Request challenge message
    2. Sign challenge with private key (wallet-specific signing scheme)
    3. `POST /auth/verify/ethereum` - Submit signature and receive access token
  - Token storage: Persisted in `~/.itzd/config.json` after successful verification

**Session Management:**
- Stateless token-based authentication
- Token embedded in `Authorization: Bearer {token}` header for authenticated requests
- Token persistence handled by `internal/config/config.go`

## Blockchain Network Support

**EVM Chains (via go-ethereum library planned):**
- Ethereum mainnet (eip155:1)
- Polygon (eip155:137)
- Arbitrum One (eip155:42161)
- Optimism (eip155:10)
- Base / Coinbase L2 (eip155:8453)
- BNB Chain (eip155:56)
- Avalanche C-Chain (eip155:43114)
- zkSync Era (eip155:324)
- Fantom (eip155:250)
- Celo (eip155:42220)

**Non-EVM L1s (planned):**
- Bitcoin (bip122:btc)
- Solana (solana:sol)
- Ripple/XRP (xrpl:mainnet)
- Aptos (aptos:mainnet)
- NEAR Protocol (near:mainnet)
- Sui (sui:mainnet)
- Algorand (algorand:mainnet)
- Cosmos Hub (cosmos:cosmoshub-4)
- Osmosis (cosmos:osmosis-1)
- Polkadot (polkadot:91b171bb)
- Kusama (polkadot:b0a8d493)
- Tezos (tezos:mainnet)
- Cardano (cardano:mainnet)
- Hedera Hashgraph (hedera:mainnet)

**Chain ID Format:**
- CAIP-2 standard (e.g., `eip155:1` for Ethereum mainnet, `bip122:btc` for Bitcoin)

## Monitoring & Observability

**Error Tracking:**
- None - errors printed to stderr

**Logs:**
- Console output only
- Error output to `os.Stderr`
- Health check status communicated to user via CLI output

## CI/CD & Deployment

**Hosting:**
- GitHub releases (binaries distributed via GitHub release assets)
  - Download URLs: `https://github.com/afterdark/itzd/releases/latest/download/itzd-{platform}-{arch}`

**CI Pipeline:**
- None detected in repository - builds appear to be manual

**Distribution:**
- Pre-compiled binaries for macOS (arm64, amd64), Linux (amd64, arm64), Windows (amd64)
- Installation via `curl` and `sudo mv` to `/usr/local/bin/`

## Environment Configuration

**Required env vars:**
- None mandatory - all configuration via `~/.itzd/config.json`
- API endpoint defaults to `https://api.itz.agency` if not configured

**Secrets location:**
- Authentication token stored in `~/.itzd/config.json` (mode 0600)
- No external secret management (Vault, etc.)

## Webhooks & Callbacks

**Incoming:**
- None

**Outgoing:**
- None - DNS records published via itz.agency API (gRPC backend)

## API Response Types

**Resolution Response:**
- Username resolution returns wallet records with: username, address, network, namespace, CAIP-2 ID, wallet address, TXT record, DNS name, verification status, DNS sync status

**Authentication Response:**
- Challenge response: challenge string, message, expiration timestamp
- Verification response: success boolean, access token, user ID, wallet address, message

**Network Metadata:**
- List of supported networks with chain codes and CAIP-2 identifiers

---

*Integration audit: 2026-05-08*
