# Technology Stack

**Analysis Date:** 2026-05-08

## Languages

**Primary:**
- Go 1.26.2 - CLI daemon and all application logic

## Runtime

**Environment:**
- Go 1.26.2 runtime

**Package Manager:**
- go modules (go.mod/go.sum)
- Lockfile: Present (`go.sum` with pinned versions)

## Frameworks

**CLI Framework:**
- Cobra 1.10.2 - Command-line interface and subcommand structure (`github.com/spf13/cobra`)

**Configuration Management:**
- Viper 1.21.0 - Configuration file parsing and environment variable management (`github.com/spf13/viper`)
- Mapstructure 2.4.0 - Struct mapping for configuration (`github.com/go-viper/mapstructure/v2`)

**Concurrency:**
- Sourcegraph conc 0.3.1 - Structured concurrency utilities (`github.com/sourcegraph/conc`)

## Key Dependencies

**Critical:**
- `spf13/cobra` 1.10.2 - Command-line application framework driving all itzd subcommands
- `spf13/viper` 1.21.0 - Configuration management (loads from `~/.itzd/config.json`)
- `fsnotify` 1.9.0 - File system monitoring for configuration changes

**Infrastructure:**
- `go-toml/v2` 2.2.4 - TOML configuration parsing
- `afero` 1.15.0 - Filesystem abstraction layer
- `locafero` 0.11.0 - Local filesystem paths
- `gotenv` 1.6.0 - .env file loading
- `yaml/v3` 3.0.4 - YAML parsing support
- `golang.org/x/sys` 0.29.0 - System-level utilities
- `golang.org/x/text` 0.28.0 - Text encoding and locale support

## Configuration

**Environment:**
- Configuration stored in `~/.itzd/config.json` (JSON format)
- Default API endpoint: `https://api.itz.agency`
- Configuration fields: `api_endpoint`, `token` (authentication), `wallet_address`, `network`

**Build:**
- Multi-platform builds via `dist/` directory
  - `itzd-darwin-arm64` (macOS Apple Silicon)
  - `itzd-darwin-amd64` (macOS Intel)
  - `itzd-linux-amd64` (Linux x86_64)
  - `itzd-linux-arm64` (Linux ARM64)
  - `itzd-windows-amd64.exe` (Windows)

## Platform Requirements

**Development:**
- Go 1.26.2 or higher
- UNIX-like environment (macOS, Linux) or Windows

**Runtime:**
- Standalone binary (no external dependencies required)
- Network access to `https://api.itz.agency` API endpoint
- HTTP client with 15-second timeout for API requests

## External API Integration

**Primary Service:**
- itz.agency API (`https://api.itz.agency`)
  - Health checks
  - Authentication (wallet challenge/verify)
  - DNS-to-wallet resolution
  - Reverse lookup (wallet to usernames)
  - Username registration
  - Network metadata

---

*Stack analysis: 2026-05-08*
