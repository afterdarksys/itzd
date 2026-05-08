# Codebase Concerns

**Analysis Date:** 2026-05-08

## Tech Debt

### Insufficient Input Validation

**Issue:** Wallet addresses, usernames, and network codes lack client-side validation before sending to API. This creates poor user experience when invalid data is submitted.

**Files:**
- `cmd/login.go:32` — Wallet address accepted with only `strings.TrimSpace()` check, no format validation
- `cmd/register.go:31-33` — Username and wallet address trimmed but not validated against format (e.g., hex length for Ethereum)
- `cmd/whoami.go:21` — Wallet address passed directly to API without validation
- `cmd/resolve.go:25` — Username accepts any input, only lowercased

**Impact:** 
- Users get confusing API error messages instead of immediate client-side feedback
- Malformed addresses (e.g., typos in Ethereum 0x prefix) waste API calls and network bandwidth
- Bad UX: no early validation feedback during `itzd register` with 42-char Ethereum hex requirement

**Fix approach:**
- Implement basic format validators: `ValidateEthereumAddress()`, `ValidateBitcoinAddress()`, `ValidateUsername()`
- For multi-chain support, chain-specific validators in `internal/validators/validators.go`
- Validate in command handlers before API calls
- Consider regex patterns per CAIP-2 spec for network codes

### Ignored ReadString Errors in Login Flow

**Issue:** `bufio.ReadString()` errors are silently ignored, allowing nil strings to proceed.

**Files:**
- `cmd/login.go:32` — `address, _ := reader.ReadString('\n')` 
- `cmd/login.go:60` — `signature, _ := reader.ReadString('\n')`

**Impact:**
- EOF or I/O errors during interactive login silently create empty strings
- Empty string checks follow (lines 34, 62), but error context is lost
- User can't distinguish between "I didn't type anything" and "read failed"

**Fix approach:**
- Capture and check error: `address, err := reader.ReadString('\n')`
- Exit with clear error message if `err != nil` (especially EOF handling)
- Provide context: "stdin closed unexpectedly" vs "wallet address required"

### Inconsistent Error Handling in Config Commands

**Issue:** `config.go` commands fail silently with `fmt.Println()` instead of exiting with error status.

**Files:**
- `cmd/config.go:41` — `set-api` writes error to stdout, returns without exit code
- `cmd/config.go:54` — `reset` writes error to stdout, returns without exit code

**Impact:**
- Scripts relying on exit codes can't detect save failures
- User sees error message but CLI exits 0 (success), confusing automation
- Config changes may not persist but user thinks they do

**Fix approach:**
- Use `fmt.Fprintln(os.Stderr, ...)` for consistency with other commands
- Call `os.Exit(1)` on save errors
- Standardize error handling across all subcommands

### No Rate-Limit or Retry Handling

**Issue:** HTTP client has fixed 15-second timeout but no exponential backoff, retry logic, or rate-limit detection.

**Files:**
- `internal/client/client.go:25` — `http.Client{Timeout: 15 * time.Second}`
- `internal/client/client.go:29-56` — `do()` method makes single attempt, no retries

**Impact:**
- Transient network failures cause immediate user-facing errors
- Rate limits from API respond with 429 but receive no special handling
- 15-second timeout may be too aggressive for slow networks (mobile)
- No detection of rate-limit `Retry-After` headers

**Fix approach:**
- Implement exponential backoff for 5xx errors (max 3 retries)
- Check for `Retry-After` header on 429 responses and respect it
- Make timeout configurable via config or env var
- Provide clear user messaging on retries: "Rate limited, waiting 5 seconds..."

## Security Considerations

### Token Stored in Plain Text on Disk

**Issue:** Authentication token saved to `~/.itzd/config.json` with file permissions `0600` (user-readable), no encryption.

**Files:**
- `internal/config/config.go:77` — `os.WriteFile(p, data, 0600)`
- `cmd/login.go:76` — Token stored directly as `cfg.Token = auth.AccessToken`

**Current mitigation:** 
- File is readable only by owner (`0600`)
- Requires shell access or local privilege to read

**Recommendations:**
- Consider encrypting token at rest: use `crypto/aes` + salt stored separately
- Alternative: Store as session cookie in system keyring (macOS Keychain, Linux Secret Service, Windows Credential Manager)
- Document risk in README: "Do not share machine access — config contains auth token"
- Add `config reset` documentation with warning about token exposure

### No HTTPS Verification / TLS Pinning

**Issue:** API calls use standard HTTP client without certificate pinning or verification override detection.

**Files:**
- `internal/client/client.go:39` — `http.NewRequest()` with no custom transport
- `internal/client/client.go:25` — Standard `http.Client` with only timeout set

**Current mitigation:**
- Production API is `https://api.itz.agency` (HTTPS enforced)
- Go's stdlib enforces TLS verification by default

**Recommendations:**
- Add TLS certificate pinning for production API to defend against supply-chain/MITM attacks
- Allow override via env var for testing: `ITZ_SKIP_VERIFY` (with prominent warning)
- Consider certificate transparency logging for added assurance

### Wallet Address Used as Password Component

**Issue:** Ethereum wallet address passed through stdin and stored in config; no PII redaction in logs/output.

**Files:**
- `cmd/login.go:32-33` — Address captured from stdin
- `cmd/login.go:85` — Address echoed back in success message
- `internal/config/config.go:19-20` — Address stored in config

**Risk:** 
- If config is accidentally shared (GitHub, pastebin), wallet address is exposed
- Logs containing address reveal blockchain transaction history

**Recommendations:**
- Redact address in logs: `0x742d35...0bEb` (first 6 + last 4 chars)
- Add warning in `login` help: "Your wallet address may be publicly visible — consider using fresh address"
- Never log full signatures or challenges even in debug mode

## Performance Bottlenecks

### String Concatenation in URL Building

**Issue:** URL paths built with `+` operator instead of `path.Join()` or `fmt.Sprintf()`.

**Files:**
- `internal/client/client.go:186` — `/resolve/` + username
- `internal/client/client.go:202` — `/resolve/` + username + `/` + network
- `internal/client/client.go:218` — `/resolve/wallet/` + walletAddress

**Impact:**
- Minimal for a CLI tool, but poor practice
- If URL contains special characters (spaces, slashes in username), concatenation breaks
- No URL encoding applied

**Fix approach:**
- Use `path.Join()` or `url.Path.Join()` (Go 1.20+)
- Use `fmt.Sprintf()` for consistency
- Example: `fmt.Sprintf("/resolve/%s/%s", username, network)`

### JSON Unmarshaling Without Size Limits

**Issue:** `json.Unmarshal()` called on response bodies with no max-size check.

**Files:**
- `internal/client/client.go:54` — `io.ReadAll(resp.Body)` reads entire response without limit
- Multiple `json.Unmarshal(data, &out)` calls throughout client

**Impact:**
- Large responses (e.g., bug in API returning huge array) can consume unbounded memory
- No protection against slow-read attacks (API sends data byte-by-byte)
- CLI becomes unresponsive if API returns multi-MB response

**Fix approach:**
- Use `io.LimitReader()` before `ReadAll()`: `io.ReadAll(io.LimitReader(resp.Body, 1MB))`
- Set reasonable limits per endpoint (e.g., resolve → 10KB, list → 100KB)
- Return error if response exceeds limit

## Fragile Areas

### Hardcoded API Endpoint in Multiple Places

**Issue:** Default API endpoint `https://api.itz.agency` appears in multiple locations; changing it requires code changes.

**Files:**
- `internal/config/config.go:12` — `DefaultAPIEndpoint` constant
- `README.md:49` — Documentation examples hardcode endpoint
- `cmd/login.go:40` — Printed in user-facing message

**Risk:**
- If API migrates, code must be rebuilt
- Different versions of CLI may point to different endpoints
- No centralized documentation of canonical endpoint

**Fix approach:**
- Keep `DefaultAPIEndpoint` in `internal/config/`
- Document in `README.md` how to override: `itzd config set-api <url>`
- Consider reading from `.env` file as fallback
- Add environment variable: `ITZ_API` with clear precedence rules

### Tight Coupling Between Commands and Config

**Issue:** All commands call `loadConfig()` directly in command handlers, no dependency injection or mock support for testing.

**Files:**
- `cmd/root.go:34-41` — Global `loadConfig()` and `newClient()` functions
- Every command file uses these globals: `cmd/login.go:25`, `cmd/register.go:35`, etc.

**Impact:**
- Cannot test commands without touching real config file (`~/.itzd/config.json`)
- No way to inject test API endpoint or mock responses
- Adding telemetry/logging to client setup requires global change

**Fix approach:**
- Create config/client provider interface: `type Provider interface { Config() *Config; Client() *Client }`
- Accept provider as parameter to each command Run() function
- Default provider reads disk, test provider uses memory
- Enables table-driven command testing without file I/O

## Error Handling

### Generic HTTP Error Messages

**Issue:** HTTP error responses often just printed as raw JSON without parsing or context.

**Files:**
- `internal/client/client.go:161` — `return nil, fmt.Errorf("challenge failed (%d): %s", code, data)`
- `internal/client/client.go:194` — `return nil, fmt.Errorf("resolve failed (%d): %s", code, data)`
- Multiple instances where `data` (raw JSON) is dumped to user

**Impact:**
- User sees `{"detail": "Invalid username"}` instead of friendly error
- No structured error handling or error type discrimination
- Hard to add retry logic specific to certain error types

**Fix approach:**
- Define error types: `type APIError struct { Status int; Code string; Detail string }`
- Unmarshal error responses into `APIError` structure
- Return meaningful text to user, not raw API response

### Silent Failures in JSON Parsing

**Issue:** `json.Unmarshal()` errors are ignored or not checked properly.

**Files:**
- `internal/client/client.go:151` — `return out, json.Unmarshal(data, &out)` (out is uninitialized map)
- `internal/client/client.go:249` — `json.Unmarshal(data, &errResp)` error ignored when parsing error response
- `internal/client/client.go:269` — Check command uses uninitialized map for response

**Impact:**
- Corrupted API responses silently return zero values
- User sees empty output instead of error message
- Difficult to debug: was API error or parsing error?

**Fix approach:**
- Always check `json.Unmarshal()` error: `if err := json.Unmarshal(...); err != nil { return nil, err }`
- Return wrapped error with context: `fmt.Errorf("parse response: %w", err)`
- Add optional verbose flag to dump raw response on parse failure

## Known Limitations

### No Support for Hardware Wallets

**Issue:** Login flow hardcoded for Ethereum EIP-191 signature only, assumes software wallet or manual signing.

**Files:**
- `cmd/login.go:41` — `api.Challenge(address, "ethereum")` — only Ethereum
- `cmd/login.go:69` — `api.VerifyEthereum()` — hardcoded chain
- `internal/client/client.go:168-182` — Only Ethereum verify endpoint

**Impact:**
- Users must manually sign or use MetaMask; no hardware wallet support
- Multi-chain users must re-authenticate separately per chain

**Recommendations:**
- Future: Support for Ledger Live / eth_signMessage via WebSocket or IPC
- Alternative: QR code flow for air-gapped hardware wallets
- Consider CAIP-122 (Sign-In with Ethereum) standard for extensibility

### Network Code Hardcoded in Commands

**Issue:** Network/chain code validation is absent; any string accepted and passed to API.

**Files:**
- `cmd/register.go:32` — Network code lowercased but not validated
- `cmd/resolve.go:30` — Network flag passed directly without validation

**Impact:**
- Typos (e.g., `ee` instead of `e`) create confusing API errors
- No client-side validation of supported networks
- Networks list in `cmd/networks.go` is static, not fetched from API

**Fix approach:**
- Cache networks list from API: `GET /networks` → stored in config with TTL
- Validate network code before API calls
- Add `--list-networks` or `-n ?` to show available codes
- Consider NetworksResp from API (already defined but not used)

### No Unauthenticated Operations Documented

**Issue:** Some operations work without login (resolve, whoami, check, networks) but others fail silently requiring `requireAuth()`.

**Files:**
- `cmd/login.go` — requires auth
- `cmd/register.go:36` — requires auth
- `cmd/resolve.go:28` — does NOT require auth
- `cmd/whoami.go:24` — does NOT require auth
- `cmd/check.go:20` — does NOT require auth

**Impact:**
- Confusing UX: some commands work without login, others don't
- README doesn't clearly document which commands need auth
- No mechanism to provide per-command auth hints

**Fix approach:**
- Document in README: "No auth required for resolve/whoami/check/networks"
- Modify `requireAuth()` to print helpful hint when missing token
- Add `--help` text indicating auth requirements per command

## Test Coverage Gaps

### No CLI Command Tests

**Issue:** No test files exist for command implementations. All logic in `cmd/*.go` is untested.

**Files:**
- No `cmd/*_test.go` files found
- No integration test harness for multi-command flows

**Risk:**
- Regression in user-facing commands discovered only in manual testing
- Refactoring (e.g., adding validation) risks breaking edge cases
- No coverage for error paths (API errors, network failures, malformed config)

**Priority:** High - User-facing commands are most critical path

**Recommendations:**
- Create `cmd/cmd_test.go` with table-driven test suite
- Mock API client for deterministic testing
- Test: happy path, missing args, API errors, network timeouts
- Use `bytes.Buffer` for stdin/stdout capture

### No Client Library Tests

**Issue:** HTTP client and JSON unmarshaling untested; all validation happens at runtime.

**Files:**
- No `internal/client/client_test.go`
- HTTP status code handling assumptions not verified

**Risk:**
- Status code edge cases (429, 402, etc.) may not behave as expected
- JSON struct tags drift from API without notice

**Recommendations:**
- Create mock HTTP server in tests
- Test each client method: Challenge, Verify, Resolve, Register, etc.
- Verify status code → error type mapping
- Test malformed JSON response handling

### No Config Persistence Tests

**Issue:** Config save/load untested; file I/O assumes success.

**Files:**
- No `internal/config/config_test.go`

**Risk:**
- Filesystem errors (permission denied, disk full) cause panics
- Config format changes break on upgrade
- JSON marshaling/unmarshaling bugs not caught

**Recommendations:**
- Mock filesystem or use `os.TempDir()` for test config
- Test: load missing file, corrupted JSON, permission errors
- Verify file permissions (0600) enforced after save

---

*Concerns audit: 2026-05-08*
