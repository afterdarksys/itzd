<!-- refreshed: 2026-05-08 -->
# Architecture

**Analysis Date:** 2026-05-08

## System Overview

```text
┌─────────────────────────────────────────────────────────────┐
│                   itzd CLI (Go Binary)                      │
├──────────────────┬──────────────────┬───────────────────────┤
│   Command Layer  │  Client Layer    │  Config Layer         │
│  `cmd/*.go`      │  `internal/client`  `internal/config`    │
│  (8 commands)    │                  │                       │
└────────┬─────────┴────────┬─────────┴──────────┬────────────┘
         │                  │                     │
         │                  ▼                     │
         │         ┌────────────────────┐        │
         │         │  HTTP Client       │        │
         │         │  (`client.Client`) │        │
         └────────▶│                    │◀───────┘
                   └────────┬───────────┘
                            │
                            ▼
                   ┌────────────────────┐
                   │  itz.agency API    │
                   │  https://api.      │
                   │  itz.agency        │
                   └────────────────────┘
                            │
                            ▼
                   ┌────────────────────┐
                   │  DNS Infrastructure│
                   │  (dnsscienced ns)  │
                   │  _waddr TXT records│
                   └────────────────────┘
```

## Component Responsibilities

| Component | Responsibility | File |
|-----------|----------------|------|
| Root Command | CLI entry point, loads config, creates client | `cmd/root.go` |
| Resolve Command | Lookup username → wallet address(es) | `cmd/resolve.go` |
| Register Command | Register username → wallet mapping | `cmd/register.go` |
| Login Command | Wallet-based authentication via signature | `cmd/login.go` |
| Whoami Command | Reverse lookup: wallet → usernames | `cmd/whoami.go` |
| List Command | Show user's registered usernames | `cmd/list.go` |
| Check Command | Availability check for username | `cmd/check.go` |
| Networks Command | Display all 24+ supported blockchain networks | `cmd/networks.go` |
| Status Command | Health check of API endpoint | `cmd/status.go` |
| Config Command | Manage local config (show, set-api, reset) | `cmd/config.go` |
| HTTP Client | API request/response handling, JWT auth | `internal/client/client.go` |
| Config Manager | Load/save ~/.itzd/config.json | `internal/config/config.go` |

## Pattern Overview

**Overall:** Thin CLI client with centralized HTTP API communication

**Key Characteristics:**
- Single binary with 9 subcommands dispatched by Cobra
- Stateless: each command loads config, creates client, makes HTTP request
- Authentication: JWT tokens stored in `~/.itzd/config.json` (mode 0600)
- Error handling: early exit (os.Exit) on validation or API failures
- No persistence: all state stored in itz.agency backend

## Layers

**CLI Layer:**
- Purpose: User-facing command interface via Cobra
- Location: `cmd/`
- Contains: Command definitions, flag parsing, output formatting
- Depends on: `internal/client`, `internal/config`
- Used by: Direct invocation by user

**Client Layer:**
- Purpose: HTTP API abstraction for itz.agency endpoints
- Location: `internal/client/client.go`
- Contains: Request building, JSON marshaling, error interpretation
- Depends on: `net/http`, `encoding/json`
- Used by: All `cmd/*` commands

**Config Layer:**
- Purpose: Local configuration persistence
- Location: `internal/config/config.go`
- Contains: Config struct, file I/O, defaults
- Depends on: `os`, `encoding/json`
- Used by: Root command, all subcommands

## Data Flow

### Primary Request Path (Resolve)

1. User invokes `itzd resolve <username>` (`cmd/resolve.go:24`)
2. Root command loads config from `~/.itzd/config.json` (`cmd/root.go:34`)
3. Root command creates HTTP client with API endpoint and token (`cmd/root.go:44`)
4. Resolve command normalizes username input (lowercase, strip `.itz.agency` suffix)
5. Client makes GET request to `/resolve/<username>` endpoint (`client/client.go:186`)
6. Response unmarshaled into `ResolveResp` struct (`client/client.go:197`)
7. Output formatted to stdout with wallet verification status and DNS info (`cmd/resolve.go:46-50`)

### Register Flow

1. User invokes `itzd register <username> <network> <wallet>` (`cmd/register.go:30`)
2. Root command loads config, checks authentication token (`cmd/root.go:49`)
3. Client makes POST request to `/username/register` with `RegisterReq` (`client/client.go:234`)
4. On 201 status, response unmarshaled to `UsernameResp`
5. Output shows registration details, DNS name, sync status
6. User can verify with `dig TXT _waddr.<username>.itz.agency` or `itzd resolve`

### Authentication Flow

1. User invokes `itzd login` (`cmd/login.go:24`)
2. User provides Ethereum wallet address via stdin
3. Client requests challenge from `/auth/challenge` endpoint (`client/client.go:156`)
4. Challenge message displayed to user
5. User signs message with wallet (cast, MetaMask, etc.)
6. Client submits signature to `/auth/verify/ethereum` (`client/client.go:169`)
7. API returns JWT access token
8. Token saved to config file with mode 0600 (`cmd/login.go:76`, `config.go:77`)

### Reverse Lookup Flow

1. User invokes `itzd whoami <wallet-address>`
2. Client makes GET request to `/resolve/wallet/<address>` (`client/client.go:218`)
3. Returns `ReverseResp` with list of usernames for that wallet

**State Management:**
- No in-memory state: all operations are stateless HTTP requests
- Persistent state: JWT token + wallet address stored in `~/.itzd/config.json`
- API owns all truth about username→wallet mappings

## Key Abstractions

**Client.do():**
- Purpose: Generic HTTP request builder with automatic header management
- Location: `internal/client/client.go:29`
- Pattern: Private helper method for GET/POST wrappers
- Handles: Content-Type, Authorization header (Bearer token), error responses

**Response Types:**
- Purpose: Typed representations of API responses
- Examples: `WalletRecord`, `ResolveResp`, `ChallengeResp`, `AuthResp`
- Pattern: Struct tags with JSON unmarshaling
- Location: `internal/client/client.go:66-141`

**Cobra Commands:**
- Purpose: CLI command hierarchy and flag binding
- Examples: Each `*Cmd` variable in `cmd/` files
- Pattern: Command.Run closure captures command logic
- Location: `cmd/root.go:13` (root), `cmd/*/go` (subcommands)

## Entry Points

**itzd Binary:**
- Location: `main.go`
- Triggers: Invoked by user (`itzd <subcommand>`)
- Responsibilities: Imports `cmd` package, calls `cmd.Execute()` to start Cobra

**cmd.Execute():**
- Location: `cmd/root.go:26`
- Triggers: Called by `main()`
- Responsibilities: Runs Cobra command tree, handles root command initialization

## Architectural Constraints

- **Threading:** Single-threaded CLI — no goroutines, all I/O sequential
- **Global state:** Command line flags stored in package-level vars (`resolveNetwork` in `resolve.go:12`), mutable config stored in `~/.itzd/config.json` with file-based locking
- **Circular imports:** None detected — dependency graph: `cmd/` → `internal/client` + `internal/config` (no reverse)
- **Error handling:** No panic recovery — early exit (os.Exit) on all errors
- **HTTP Timeout:** Fixed 15-second timeout for all API calls (`client/client.go:25`)
- **Configuration:** Single source of truth is `~/.itzd/config.json` (XDG not used)

## Anti-Patterns

### Repeated Config Loading

**What happens:** Each command calls `loadConfig()` independently, re-reading the same file from disk
**Why it's wrong:** Inefficient for multi-command scripts, inconsistent view if config changes mid-session
**Do this instead:** Load config once in `Execute()`, pass to all subcommands via context

Example in `cmd/resolve.go:27`: `cfg := loadConfig()` inside the Run closure

### No Structured Logging

**What happens:** All output via `fmt.Println` and `fmt.Fprintf(os.Stderr, ...)`, mixed with user-facing text
**Why it's wrong:** Cannot distinguish informational output from errors programmatically, no debug mode
**Do this instead:** Use a logging library (e.g., `log/slog`) with levels and structured fields

Example in `cmd/register.go:39`: `fmt.Printf("Registering %s.itz.agency...` and error output mixed at same level

### Magic String Parsing

**What happens:** Username normalization via `strings.ToLower`, `strings.TrimSpace`, `strings.TrimSuffix` manually applied in each command
**Why it's wrong:** Inconsistent normalization logic scattered across `resolve.go`, `register.go`, `check.go`, `login.go`
**Do this instead:** Centralize in a `NormalizeUsername()` function in `internal/` package

Example in `cmd/resolve.go:25` vs `cmd/register.go:31`: Different approaches to trimming suffix

## Error Handling

**Strategy:** Fast-fail with explicit error messages to stderr, exit with code 1

**Patterns:**
- HTTP errors: Check status code, format error message with code and response body (`client/client.go:161`)
- Auth errors: Exit with "not logged in" hint (`cmd/root.go:51`)
- File I/O errors: Wrap with context (`config/config.go:51`)
- JSON parsing errors: Pass raw error with context (`config/config.go:55`)

Example (`cmd/register.go:41-44`):
```go
rec, err := api.Register(username, network, wallet)
if err != nil {
    fmt.Fprintln(os.Stderr, "error:", err)
    os.Exit(1)
}
```

## Cross-Cutting Concerns

**Logging:** Console output only, no structured logging library. Status messages to stdout, errors to stderr.

**Validation:** 
- Username: normalized to lowercase, suffix stripped
- Network: passed as-is to API, validated by backend
- Wallet address: passed as-is, validated by backend
- JWT token: accepted as-is from API, stored in config

**Authentication:** 
- Mechanism: Wallet signature challenge-response, JWT token storage
- Scope: Token grants access to user's own usernames
- Token storage: `~/.itzd/config.json` with file permissions 0600

---

*Architecture analysis: 2026-05-08*
