# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

VoidVPN is a cross-platform CLI VPN client written in Go 1.22. It supports two protocols: embedded WireGuard (via `wireguard-go`, no external tools needed) and OpenVPN (shells out to the `openvpn` binary). The primary target is Windows, with Linux and macOS also supported.

## Build & Development Commands

```bash
make build                # Build for current platform (output: ./voidvpn or voidvpn.exe)
make test                 # Run all tests: go test ./... -v -count=1 -timeout 120s
make vet                  # Run go vet ./...
make fmt                  # Format with gofmt -s -w .
make lint                 # Run golangci-lint (must be installed separately)
make test-coverage        # Tests with coverage report → coverage.out
make build-all            # Cross-compile for windows/linux/darwin amd64
make install              # Install to $GOPATH/bin with ldflags
make docker-build         # Build Docker image with version tags
make completions          # Generate shell completions (bash/zsh/fish/powershell)
make clean                # Remove build artifacts and coverage.out
```

Run a single test:
```bash
go test ./internal/config/ -v -run TestConfigLoadSave
```

Version info is injected via ldflags (`pkg/version`). The Makefile sets `-X` flags for Version, Commit, and BuildDate automatically.

On Windows, `wintun.dll` must be alongside the binary at runtime.

CI runs tests on a 3-platform matrix (ubuntu, windows, macos) with Go 1.22. Build matrix covers amd64+arm64 for all three OS targets. Releases are triggered by `v*` tags.

## Architecture

### Protocol Abstraction

The `internal/tunnel` package defines the `Tunnel` interface (`Connect`, `Disconnect`, `Status`, `IsActive`). Two implementations exist:
- `internal/wireguard` — Embedded WireGuard using `wireguard-go`. Creates TUN device, configures via IPC protocol, runs in-process. Private key is loaded from keystore at connect time.
- `internal/openvpn` — Shells out to the `openvpn` binary. Writes a temp config, launches the process, and monitors stdout/stderr for "Initialization Sequence Completed" or error strings. Uses the OpenVPN management interface (TCP on a random port 10000-60000) for stats and graceful shutdown.

The CLI (`internal/cli/connect.go`) selects which tunnel to instantiate based on `serverCfg.Protocol` (`"openvpn"` → OpenVPN, anything else → WireGuard).

### Connect Flow

The connect command uses an async pattern: `daemon.Run()` starts in a goroutine, a bubbletea spinner shows progress, and `daemon.Connected` (a channel) signals success. The spinner receives a `ui.ConnectMsg` on either connection success or failure. After the spinner exits, the main goroutine blocks on the daemon error channel until Ctrl+C.

### Platform-Specific Code

Platform splits use build tags (`//go:build windows` / `//go:build !windows`) and file suffixes (`_windows.go`, `_unix.go`, `_linux.go`). Key splits:

| Concern | Windows | Unix |
|---------|---------|------|
| Admin check | Windows SID / token membership | `os.Geteuid() == 0` |
| TUN creation | Wintun (extracts LUID) | Direct `tun.CreateTUN` |
| Daemon IPC | TCP localhost:41820 + hex token auth | Unix domain socket + FS permissions |
| DNS | `netsh` commands | `/etc/resolv.conf` rewrite |
| Routes | `route.exe` commands | `ip route` commands |

### Daemon & IPC

`internal/daemon` orchestrates the tunnel lifecycle: connect → configure DNS/routes → start IPC server → wait for signal → cleanup. The `Connected` channel signals async connection success to the CLI layer. State is persisted as JSON in the OS config directory.

IPC uses JSON messages (`IPCRequest`/`IPCResponse`). Windows uses TCP with a random token stored in `state/ipc.token`. Unix uses a domain socket with filesystem-level auth.

### Keystore

`internal/keystore` provides a `Keystore` interface with two backends: OS keyring (via `zalando/go-keyring`) and an AES-256-GCM encrypted file fallback. `New()` probes the keyring at init and falls back silently. Key names are validated against `^[a-zA-Z0-9][a-zA-Z0-9_-]{0,62}$` to prevent path traversal.

### UI

`internal/ui` uses the Charmbracelet stack (lipgloss, bubbletea, bubbles). Brand colors: purple `#7B2FBE` (primary), cyan `#00D4FF` (accent). The connect flow runs a bubbletea program with a spinner model that receives `ConnectMsg` on connection success/failure.

### CLI

Built with Cobra (`spf13/cobra`). Root command in `internal/cli/root.go` registers all subcommands in `init()`. `PersistentPreRun` ensures config directories exist. Global `--verbose` flag.

### Config

YAML-based, stored in OS-standard config dirs (`%APPDATA%\VoidVPN\` on Windows, `~/.config/voidvpn/` on Linux). Each server is a separate YAML file under `servers/`. Runtime connection state is JSON under `state/`.

The `ServerConfig` struct has a `Protocol` field that determines tunnel type. Import supports both WireGuard `.conf` files (parsed via `gopkg.in/ini.v1`) and OpenVPN `.ovpn` files (line-by-line directive parsing with inline block extraction for certs/keys).

## Key Dependencies

- `golang.zx2c4.com/wireguard` — WireGuard protocol implementation
- `golang.zx2c4.com/wireguard/windows` — Windows-specific WireGuard/Wintun support
- `github.com/spf13/cobra` — CLI framework
- `github.com/charmbracelet/{lipgloss,bubbletea,bubbles}` — Terminal UI
- `github.com/zalando/go-keyring` — OS credential storage
- `gopkg.in/yaml.v3` — Config serialization
- `gopkg.in/ini.v1` — WireGuard .conf file parsing

## Testing Patterns

Standard `testing` package with table-driven tests. No mocking framework — tests use real filesystem via `t.TempDir()`. Environment variables (e.g., `APPDATA`) are overridden in tests for path isolation.

## Security Considerations

All shell-executed values (DNS IPs, route CIDRs, interface names) are validated before being passed to system commands to prevent injection. The WireGuard IPC config builder in `internal/wireguard/ipc.go` validates that no newlines appear in endpoint/AllowedIPs strings. Routes use the `0.0.0.0/1` + `128.0.0.0/1` split strategy to avoid deleting the default route.
