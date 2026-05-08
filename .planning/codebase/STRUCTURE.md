# Codebase Structure

**Analysis Date:** 2026-05-08

## Directory Layout

```
itzd/
├── cmd/                    # CLI command implementations
│   ├── root.go            # Root command, config/client setup
│   ├── login.go           # Wallet authentication
│   ├── register.go        # Register username → wallet
│   ├── resolve.go         # Lookup username → wallets
│   ├── whoami.go          # Reverse lookup: wallet → usernames
│   ├── list.go            # List user's registered names
│   ├── check.go           # Check username availability
│   ├── networks.go        # Display supported chains
│   ├── config.go          # Config management (show, set-api, reset)
│   └── status.go          # Health check API endpoint
├── internal/              # Private packages
│   ├── client/            # HTTP API client
│   │   └── client.go      # Request building, response types
│   └── config/            # Configuration management
│       └── config.go      # Load/save ~/.itzd/config.json
├── main.go                # Entry point (calls cmd.Execute)
├── go.mod                 # Go module definition
├── go.sum                 # Dependency checksums
├── README.md              # User documentation
├── itzd                   # Compiled binary (built via `go build`)
└── dist/                  # Release binaries (macOS, Linux)
```

## Directory Purposes

**cmd/:**
- Purpose: All 9 subcommands for the itzd CLI
- Contains: Cobra command definitions, flag parsing, output formatting
- Key files: `root.go` (orchestrates), `login.go` (auth), `register.go` (registration), `resolve.go` (lookup)

**internal/client/:**
- Purpose: HTTP API client abstraction
- Contains: `Client` struct, all API method signatures, request/response types
- Key files: `client.go` (only file, contains ~284 lines of types and methods)

**internal/config/:**
- Purpose: Local configuration management
- Contains: `Config` struct, file I/O, defaults
- Key files: `config.go` (only file, ~79 lines)

**dist/:**
- Purpose: Compiled release binaries
- Contains: Pre-built binaries for darwin-arm64, darwin-amd64, linux-amd64
- Generated: By build pipeline, not source

## Key File Locations

**Entry Points:**
- `main.go`: Single entry point that imports and calls `cmd.Execute()`

**Configuration:**
- `internal/config/config.go`: Manages `~/.itzd/config.json` (user's home directory)

**Core Logic:**
- `cmd/root.go`: Config loading, client creation, auth checking (used by all subcommands)
- `internal/client/client.go`: All API endpoints and request/response marshaling

**Commands (alphabetical):**
- `cmd/check.go`: Username availability check (`itzd check <username> <network>`)
- `cmd/config.go`: Config management (`itzd config {show,set-api,reset}`)
- `cmd/list.go`: List user's usernames (`itzd list`)
- `cmd/login.go`: Wallet authentication (`itzd login`)
- `cmd/networks.go`: Show supported chains (`itzd networks`)
- `cmd/register.go`: Register username (`itzd register <username> <network> <wallet>`)
- `cmd/resolve.go`: Lookup username (`itzd resolve <username>`)
- `cmd/status.go`: Health check (`itzd status`)
- `cmd/whoami.go`: Reverse lookup (`itzd whoami <wallet>`)

## Naming Conventions

**Files:**
- Go source: `*.go` (lowercase, dash-separated for multi-word: none currently)
- Package folders: lowercase (e.g., `cmd/`, `internal/client/`)

**Directories:**
- Package name matches directory name
- Standard Go convention: `cmd/` for commands, `internal/` for private packages

**Functions:**
- Public (exported): PascalCase (e.g., `New()`, `Execute()`, `Load()`, `Save()`)
- Private (package-scoped): camelCase (e.g., `loadConfig()`, `requireAuth()`, `newClient()`)
- Examples from `cmd/root.go:34-46`: `loadConfig()`, `newClient()`, `requireAuth()`

**Variables:**
- Package-level flags: camelCase (e.g., `resolveNetwork` in `resolve.go:12`)
- Struct fields: PascalCase (e.g., `Token`, `APIEndpoint`, `WalletAddress`)

**Types:**
- Structs: PascalCase (e.g., `Client`, `Config`, `WalletRecord`, `ResolveResp`)
- Interfaces: PascalCase ending in "-er" if methods (not applicable here)
- Examples: `client/client.go:14-17` defines `Client` struct

## Where to Add New Code

**New Command:**
- Implementation: Create `cmd/newcommand.go` with `var newcommandCmd = &cobra.Command{...}`
- Register: Add command to `rootCmd.AddCommand()` in `cmd/root.go:57-67`
- Pattern: Follow structure of `cmd/resolve.go` or `cmd/register.go`
- Example: To add `itzd export`, create `cmd/export.go` and add `exportCmd` to init in root.go

**New API Endpoint:**
- HTTP method: Add method to `Client` struct in `internal/client/client.go`
- Pattern: Use existing `do()`, `get()`, `post()` helpers (lines 29-64)
- Request type: Add request struct near top (e.g., `RegisterReq` at line 120)
- Response type: Add response struct (e.g., `UsernameResp` at line 126)
- Example: To add `/api/v2/custom` endpoint:
  ```go
  // In internal/client/client.go
  type CustomReq struct { ... }
  type CustomResp struct { ... }
  func (c *Client) Custom(req *CustomReq) (*CustomResp, error) {
      data, code, err := c.post("/custom", req)
      // handle error, unmarshal
  }
  ```

**Shared Utilities:**
- Location: Create new file in `internal/` (e.g., `internal/utils/utils.go`)
- Pattern: Exported functions in `internal/` package, imported by `cmd/`
- Example: To centralize username normalization, create `internal/normalize/normalize.go:NormalizeUsername()`

**Tests:**
- Not currently in repo
- If added: Use `*_test.go` naming convention (e.g., `cmd/resolve_test.go`)
- Pattern: Table-driven tests for commands, unit tests for `client.go` and `config.go`

## Special Directories

**dist/:**
- Purpose: Compiled release binaries
- Generated: Yes (by build process)
- Committed: No (binaries not in source control)

**\.git/:**
- Purpose: Version control metadata
- Generated: Yes
- Committed: N/A (special directory)

**\.planning/:**
- Purpose: Codebase documentation (this directory)
- Generated: No (manually created by GSD mapping)
- Committed: Yes (reference documentation)

## Go Module Configuration

**Module:** `itz.agency/itzd`
**Go Version:** 1.26.2
**Key Dependencies:**
- `github.com/spf13/cobra` v1.10.2 — CLI framework
- `github.com/spf13/viper` v1.21.0 — Configuration management (indirectly via Cobra)

## Configuration Files

**go.mod/go.sum:**
- Location: Root directory
- Purpose: Define dependencies and versions

**~/.itzd/config.json (user's home):**
- Purpose: Persistent user configuration
- Created by: `config.Save()` after login
- Structure: JSON with `api_endpoint`, `token`, `wallet_address`, `network`
- Permissions: 0600 (readable/writable by owner only)

---

*Structure analysis: 2026-05-08*
