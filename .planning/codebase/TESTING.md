# Testing Patterns

**Analysis Date:** 2026-05-08

## Test Framework

**Status:** Not detected

**Current State:**
- No `*_test.go` files found in codebase
- No test configuration files detected (`go.test.v`, `testify`, `gotest.tools` imports)
- No testing framework dependencies visible in `go.mod`
- Testing is not currently implemented in this codebase

**Typical Go Testing Setup (when implemented):**
- Runner: `go test ./...` (built into Go)
- Assertion Library: None detected; uses standard library testing package or third-party library like `testify/assert`
- Run Commands:
```bash
go test ./...              # Run all tests
go test -v ./...           # Verbose output
go test -cover ./...       # Coverage report
go test -race ./...        # Race condition detection
```

## Code Organization for Future Testing

**Current File Structure:**
- Commands: `cmd/*.go` (e.g., `cmd/resolve.go`, `cmd/login.go`)
- Config: `internal/config/config.go`
- Client: `internal/client/client.go`
- Main entry: `main.go` → `cmd.Execute()`

**Test File Location Pattern (when added):**
- Test files would use `*_test.go` naming convention
- Located in same package as source: `cmd/resolve_test.go`, `internal/client/client_test.go`
- Example structure:
```
cmd/
  resolve.go
  resolve_test.go      # Tests for resolve command
  login.go
  login_test.go        # Tests for login command
  root.go
  root_test.go         # Tests for root helpers

internal/
  config/
    config.go
    config_test.go     # Tests for config load/save
  client/
    client.go
    client_test.go     # Tests for HTTP client
```

## Areas Requiring Test Coverage

**High Priority (API interactions):**
- `internal/client/client.go` - HTTP client methods:
  - `Challenge()` - Challenge request handling
  - `VerifyEthereum()` - Signature verification
  - `Resolve()` - Username resolution
  - `ResolveNetwork()` - Single network resolution
  - `Whoami()` - Reverse lookup
  - `Register()` - Username registration
  - `MyUsernames()` - Listing usernames
  - Error code handling (404, 409, 402, 200/201)

**Medium Priority (Configuration):**
- `internal/config/config.go` - Config lifecycle:
  - `Load()` - File reading, defaults, missing file handling
  - `Save()` - File writing, directory creation with permissions
  - Default API endpoint assignment
  - Config parsing errors

**Medium Priority (Commands):**
- `cmd/resolve.go` - Argument parsing, flag handling
- `cmd/login.go` - Interactive input simulation, token storage
- `cmd/register.go` - CLI argument validation
- Command initialization and Cobra integration

**Lower Priority (Display):**
- Output formatting helpers: `printRecord()`, `plural()`
- Display functions that don't affect system state

## Mocking Strategy (when implementing)

**External Dependencies to Mock:**
- `net/http.Client` - HTTP requests in `internal/client/client.go`
- `os.File` operations - File I/O in `internal/config/config.go`
- `os.UserHomeDir()` - Directory resolution in config
- User stdin - Interactive prompts in `cmd/login.go`

**Mocking Approach:**
- Use dependency injection: Pass `*http.Client` to `Client` constructor rather than creating inline
- Use interface types: Define `Reader` interface for config file reading
- Use `httptest.NewServer()` for testing HTTP client against mock API

**Example Mock Pattern (when implemented):**
```go
// In tests, inject mock HTTP client
mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    // Mock response
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(expectedResp)
}))

client := &Client{
    base:  mockServer.URL,
    token: "test-token",
    http:  mockServer.Client(),
}
```

## Test Structure Pattern (when implemented)

**Standard Go test pattern:**
```go
package client

import (
    "testing"
)

func TestResolve(t *testing.T) {
    // Setup
    client := newTestClient()
    
    // Execute
    result, err := client.Resolve("username")
    
    // Assert
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if result.Count != 3 {
        t.Errorf("expected 3 wallets, got %d", result.Count)
    }
}

func TestResolveNotFound(t *testing.T) {
    client := newTestClient()
    
    result, err := client.Resolve("nonexistent")
    
    if err == nil {
        t.Fatal("expected error for nonexistent username")
    }
    if result != nil {
        t.Errorf("expected nil result, got %v", result)
    }
}
```

## Error Handling in Tests

**Pattern to verify:**
- Commands exit with status code 1 on error: `os.Exit(1)`
- API client returns `(*Type, error)` tuples
- Config operations return error on file I/O failure
- Status code checks before unmarshaling JSON

**Test cases needed:**
- Network errors (timeout, connection refused)
- HTTP error status codes (404, 409, 402, 500)
- Invalid JSON responses
- Missing environment/config
- Authentication failures

## Coverage Targets (when implementing)

**Recommended minimum coverage:** 70% overall

**Coverage by module:**
- `internal/client/client.go`: 80%+ (critical API interaction layer)
- `internal/config/config.go`: 75%+ (system configuration)
- `cmd/` package: 50%+ (CLI interactions are harder to test)

**View coverage (when tests exist):**
```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out        # View in browser
go tool cover -func=coverage.out        # Summary by function
```

## HTTP Status Code Testing

**Codes to handle (in `internal/client/client.go`):**
- 200: Success (GET, most POST operations)
- 201: Created (Register endpoint)
- 404: Not found (username doesn't exist)
- 409: Conflict (username already taken)
- 402: Payment required (plan limit exceeded)
- 5xx: Server errors

**Test cases:**
```go
// Success case
func TestResolveSuccess(t *testing.T) {
    // Mock server returns 200 with ResolveResp
    // Verify unmarshaling succeeds
}

// 404 case
func TestResolveNotFound(t *testing.T) {
    // Mock server returns 404
    // Verify error message includes "not found"
}

// 409 case (Register)
func TestRegisterConflict(t *testing.T) {
    // Mock server returns 409
    // Verify error message includes "already taken"
}

// 402 case (Register)
func TestRegisterPlanLimit(t *testing.T) {
    // Mock server returns 402 with detail.message
    // Verify plan limit error is extracted
}
```

## Integration Test Considerations (when scaling)

**Current blockers:**
- No test fixtures or test data defined
- No test configuration separate from production
- No mock API server or test endpoint

**When needed (future phases):**
- Use `httptest.NewServer()` for integration tests
- Separate test API endpoint configuration
- Create test fixtures for config files
- Consider testcontainers for database testing (if database added later)

---

*Testing analysis: 2026-05-08*
