# Coding Conventions

**Analysis Date:** 2026-05-08

## Naming Patterns

**Files:**
- `snake_case.go` for Go source files
- Command files: `<verb>.go` pattern (e.g., `resolve.go`, `register.go`, `login.go`)
- Package organization: `cmd/` for commands, `internal/<package>/` for private packages
- Example: `cmd/resolve.go`, `internal/config/config.go`, `internal/client/client.go`

**Functions:**
- `camelCase` starting with lowercase for unexported functions
- `PascalCase` starting with uppercase for exported functions
- Verb-first naming: `Resolve()`, `ResolveNetwork()`, `Challenge()`, `VerifyEthereum()`, `Register()`, `Whoami()`
- Helper functions use lowercase: `loadConfig()`, `newClient()`, `requireAuth()`, `printRecord()`, `plural()`
- Example pattern: `func (c *Client) Resolve(username string) (*ResolveResp, error)` in `internal/client/client.go`

**Variables:**
- `camelCase` for local and package-level variables
- Short names for temporary variables: `err` for errors, `rec` for records, `cfg` for config, `api` for API client
- Receiver names: `c` for Client, `cmd` for Cobra Command
- Example: `cfg := loadConfig()`, `api := newClient(cfg)`, `rec, err := api.ResolveNetwork(...)`

**Types:**
- `PascalCase` for exported struct names: `Config`, `Client`, `WalletRecord`, `ResolveResp`, `ChallengeReq`
- `PascalCase` for exported field names with JSON struct tags for serialization
- Response/Request types: `<Action>Resp` and `<Action>Req` pattern (e.g., `ChallengeResp`, `ChallengeReq`, `VerifyReq`, `AuthResp`)
- Example: `type ResolveResp struct { Username string; Wallets []WalletRecord; Count int }`

## Code Style

**Formatting:**
- Standard Go gofmt (automatically enforced by language)
- 2-space indentation (Go standard)
- Lines stay readable (no hard limit enforced, but typically under 100 characters)
- Example: imports organized in groups, logical line breaks between sections

**Linting:**
- No explicit `.golangci.yml` found; relies on Go standard conventions
- Follow idiomatic Go patterns (error handling, interface naming, etc.)
- No custom linting rules detected

**Comment Style:**
- Package-level comments: `// Package <name> <description>` pattern at top of file
- Example: `// Package cmd implements the itzd CLI.` in `cmd/root.go`
- Function comments use standard format: `// <FunctionName> <description>`
- Example: `// Challenge requests an auth challenge for a wallet address.` in `internal/client/client.go`

## Import Organization

**Order:**
1. Standard library imports (fmt, os, strings, net/http, etc.)
2. Third-party imports (github.com/spf13/cobra, etc.)
3. Internal imports (itz.agency/itzd/...)

**Example from `cmd/resolve.go`:**
```go
import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"itz.agency/itzd/internal/client"
)
```

**Path Aliases:**
- No aliases detected. Uses full module path: `itz.agency/itzd/cmd`, `itz.agency/itzd/internal/config`

## Error Handling

**Patterns:**
- Standard Go error returns: `(result Type, error)`
- Two-return pattern is universal: `data, code, err := c.get(path)`
- Error checking: `if err != nil { ... }`
- On CLI, errors are logged to stderr and followed by `os.Exit(1)`
- Example from `cmd/resolve.go`:
```go
resp, err := api.Resolve(username)
if err != nil {
    fmt.Fprintln(os.Stderr, "error:", err)
    os.Exit(1)
}
```
- HTTP status codes checked after receiving data: `if code != 200 { return nil, fmt.Errorf(...) }`
- Example from `internal/client/client.go`:
```go
if code != 200 {
    return nil, fmt.Errorf("challenge failed (%d): %s", code, data)
}
```

## Logging

**Framework:** 
- Standard library `fmt` package only
- No structured logging library used

**Patterns:**
- Info messages: `fmt.Printf()`, `fmt.Println()` to stdout
- Error messages: `fmt.Fprintln(os.Stderr, ...)` to stderr
- Example: `fmt.Fprintln(os.Stderr, "error:", err)` for error output
- Success messages use Unicode symbols: `fmt.Printf("✓ Logged in as %s\n", ...)`
- Status indicators: `✓` for success, `✗` for failure, `⚠` for warnings

## Module Design

**Structure:**
- `cmd/` package: Command definitions using Cobra framework
- `internal/config/` package: Configuration loading/saving
- `internal/client/` package: HTTP API client
- Entry point: `main.go` calls `cmd.Execute()`

**Exports:**
- Commands defined at package level: `var <verb>Cmd = &cobra.Command{...}`
- Package initialization in `init()` functions: `rootCmd.AddCommand(...)`
- Example from `cmd/root.go`:
```go
func init() {
	rootCmd.AddCommand(
		configCmd,
		loginCmd,
		resolveCmd,
		// ...
	)
}
```

**Cobra Command Pattern:**
Each command file follows this structure:
```go
var <verb>Cmd = &cobra.Command{
    Use:   "verb <args>",
    Short: "One-line description",
    Long:  "Multi-line help text with examples",
    Args:  cobra.ExactArgs(n),
    Run: func(cmd *cobra.Command, args []string) {
        // Implementation
    },
}

func init() {
    <verb>Cmd.Flags().StringVarP(&flagVar, "name", "short", "", "description")
}
```

**Configuration Pattern:**
- Config loaded once per command: `cfg := loadConfig()`
- API client created from config: `api := newClient(cfg)`
- Auth check before operations: `requireAuth(cfg)`
- Example from `cmd/register.go`:
```go
cfg := loadConfig()
requireAuth(cfg)
api := newClient(cfg)
```

## HTTP Client Pattern

**Receiver Method Design:**
- Generic `do(method, path, body)` method with helper wrappers `get()` and `post()`
- All responses return `([]byte, int, error)` for raw data, status code, and error
- Example from `internal/client/client.go`:
```go
func (c *Client) do(method, path string, body any) ([]byte, int, error) {
    // HTTP request handling
}

func (c *Client) get(path string) ([]byte, int, error) {
    return c.do("GET", path, nil)
}
```

**Response Parsing:**
- Manual JSON unmarshaling after status code check
- Always verify `code == 200` or specific success code before unmarshaling
- Example:
```go
data, code, err := c.get(path)
if code != 200 {
    return nil, fmt.Errorf("failed (%d): %s", code, data)
}
var out WalletRecord
return &out, json.Unmarshal(data, &out)
```

## Helper Functions

**CLI helpers in `cmd/root.go`:**
- `loadConfig()`: Loads config with error handling and exit
- `newClient(cfg)`: Creates API client from config
- `requireAuth(cfg)`: Checks auth token, exits if missing
- Example pattern:
```go
func loadConfig() *Config {
    cfg, err := config.Load()
    if err != nil {
        fmt.Fprintln(os.Stderr, "error loading config:", err)
        os.Exit(1)
    }
    return cfg
}
```

**Display helpers:**
- `printRecord(w *client.WalletRecord)`: Formats wallet record output with symbols
- `plural(n int)`: Returns "s" suffix for plural, empty for singular
- Example: `fmt.Printf("%d wallet%s\n", resp.Count, plural(resp.Count))`

## Type Organization in Modules

**Client module (`internal/client/client.go`):**
- Request/Response types grouped together with separator comment `// ── Types ────`
- Methods grouped with separator comment `// ── Methods ──`
- Related functionality (Challenge, Verify, Resolve) grouped logically
- Example:
```go
// ── Types ────────────────────────────────────────────────────────────────────

type ChallengeReq struct { ... }
type ChallengeResp struct { ... }

// ── Methods ──────────────────────────────────────────────────────────────────

func (c *Client) Challenge(...) { ... }
```

---

*Convention analysis: 2026-05-08*
