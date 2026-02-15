# Contributing to VoidVPN

Thank you for your interest in contributing to VoidVPN. This document provides
guidelines and instructions for contributing to the project.

---

## Table of Contents

- [Development Setup](#development-setup)
- [Building](#building)
- [Testing](#testing)
- [Code Style](#code-style)
- [Branch Naming](#branch-naming)
- [Commit Message Format](#commit-message-format)
- [Pull Request Process](#pull-request-process)
- [Project Structure](#project-structure)

---

## Development Setup

### Prerequisites

- **Go 1.22+** -- download from [go.dev](https://go.dev/dl/)
- **Git** -- for version control
- **Make** (optional) -- for using Makefile targets
- **golangci-lint** (optional) -- for extended linting

### Getting Started

1. Fork the repository on GitHub.

2. Clone your fork:

   ```bash
   git clone https://github.com/<your-username>/voidvpn.git
   cd voidvpn
   ```

3. Download dependencies:

   ```bash
   go mod download
   ```

4. Verify dependencies:

   ```bash
   go mod verify
   ```

5. Run the tests to confirm everything works:

   ```bash
   go test ./... -v
   ```

### Platform Notes

VoidVPN includes platform-specific code for Windows, Linux, and macOS. Build
tags and OS-suffixed files (e.g., `_windows.go`, `_linux.go`, `_unix.go`) are
used to isolate platform-dependent logic. You can develop on any supported OS,
but be aware that some code paths will only compile on their target platform.

---

## Building

### Using go build

```bash
go build ./cmd/voidvpn/
```

This produces a `voidvpn` binary (or `voidvpn.exe` on Windows) in the current
directory.

### Using Make

```bash
make build
```

The Makefile injects version information via ldflags automatically. Additional
build targets are available:

| Target            | Description                              |
|-------------------|------------------------------------------|
| `make build`      | Build for the current OS/architecture    |
| `make build-windows` | Cross-compile for Windows (amd64)     |
| `make build-linux`   | Cross-compile for Linux (amd64)       |
| `make build-darwin`  | Cross-compile for macOS (amd64)       |
| `make build-all`     | Build for all platforms               |
| `make clean`         | Remove built binaries and coverage    |
| `make install`       | Install to $GOPATH/bin                |

### Docker

```bash
make docker-build
```

Or directly:

```bash
docker build -t voidvpn:dev .
```

---

## Testing

### Running Tests

Run the full test suite:

```bash
go test ./... -v -count=1 -timeout 120s
```

Or using Make:

```bash
make test
```

### Test Coverage

Generate a coverage report:

```bash
make test-coverage
```

This produces a `coverage.out` file and prints a per-function coverage summary.

### Writing Tests

- Place test files alongside the code they test (e.g., `config.go` and
  `config_test.go` in the same package).
- Use the standard `testing` package. Use `testify` for assertions when it
  improves readability.
- Name test functions descriptively: `TestConfigLoad_MissingFile`,
  `TestKeypairGeneration_ValidOutput`.
- Use table-driven tests where multiple input/output combinations are tested.
- Avoid external service dependencies in unit tests. Use interfaces and mocks
  for system-level operations.

---

## Code Style

VoidVPN follows standard Go conventions. All contributions must pass `gofmt`
and `go vet` without warnings.

### Formatting

Format your code before committing:

```bash
gofmt -s -w .
```

Or using Make:

```bash
make fmt
```

### Vetting

Run the Go static analyzer:

```bash
go vet ./...
```

Or using Make:

```bash
make vet
```

### Linting (optional but recommended)

If you have `golangci-lint` installed:

```bash
make lint
```

### General Guidelines

- Follow [Effective Go](https://go.dev/doc/effective_go) and the
  [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments).
- Use meaningful, descriptive names. Avoid single-letter variables except in
  short loop scopes.
- Export only what needs to be public. Keep the public API surface minimal.
- Add doc comments to all exported types, functions, and constants.
- Handle all errors explicitly. Do not discard errors with `_` unless there is
  a documented reason.
- Keep functions focused and short. If a function exceeds roughly 50 lines,
  consider splitting it.
- Use `internal/` packages to prevent external imports of implementation
  details.
- Place reusable, stable utilities in `pkg/`.

---

## Branch Naming

Use the following prefixes for branch names:

| Prefix       | Purpose                                    |
|--------------|--------------------------------------------|
| `feature/`   | New features or capabilities               |
| `fix/`       | Bug fixes                                  |
| `refactor/`  | Code restructuring without behavior change |
| `docs/`      | Documentation-only changes                 |

Examples:

```
feature/server-import-wildcard
fix/dns-restore-on-disconnect
refactor/daemon-state-machine
docs/update-contributing-guide
```

Always branch from `develop` (or `main` if there is no `develop` branch).

---

## Commit Message Format

VoidVPN uses [Conventional Commits](https://www.conventionalcommits.org/).

### Structure

```
<type>(<scope>): <description>

[optional body]

[optional footer(s)]
```

### Types

| Type         | When to use                                |
|--------------|--------------------------------------------|
| `feat`       | A new feature                              |
| `fix`        | A bug fix                                  |
| `docs`       | Documentation changes only                 |
| `style`      | Formatting, missing semicolons, etc.       |
| `refactor`   | Code change that neither fixes nor adds    |
| `test`       | Adding or updating tests                   |
| `chore`      | Build process, CI, dependency updates      |
| `perf`       | Performance improvements                   |
| `ci`         | CI/CD configuration changes                |

### Scopes

Use the package or area of the change as the scope:

`cli`, `config`, `daemon`, `keystore`, `network`, `platform`, `ui`,
`wireguard`, `docker`, `ci`

### Examples

```
feat(cli): add server import command for .conf files
fix(network): restore DNS settings on disconnect failure
refactor(daemon): extract IPC message parsing into helper
test(config): add table-driven tests for YAML parsing
docs: update contributing guide with branch naming
chore(ci): add arm64 to build matrix
```

### Rules

- Use the imperative mood in the description ("add", not "added" or "adds").
- Do not capitalize the first letter of the description.
- Do not end the description with a period.
- Keep the first line under 72 characters.
- Use the body to explain *what* and *why*, not *how*.

---

## Pull Request Process

1. **Create a branch** from `develop` (or `main`) using the naming conventions
   above.

2. **Make your changes.** Keep commits atomic and focused. Each commit should
   represent a single logical change.

3. **Ensure quality checks pass locally:**

   ```bash
   gofmt -s -w .
   go vet ./...
   go test ./... -v -count=1 -timeout 120s
   ```

4. **Push your branch** to your fork:

   ```bash
   git push origin feature/your-feature-name
   ```

5. **Open a pull request** against the `main` (or `develop`) branch of the
   upstream repository.

6. **Fill in the PR description:**
   - Summarize what the PR does and why.
   - Reference any related issues (e.g., `Closes #42`).
   - Note any breaking changes.
   - Include testing steps if the change is non-trivial.

7. **CI must pass.** The GitHub Actions workflow runs `go vet`, tests on
   Ubuntu, Windows, and macOS, and cross-compilation for all target platforms.

8. **Address review feedback.** Push additional commits to your branch to
   address reviewer comments. Do not force-push during review unless asked.

9. **Merge.** Once approved, a maintainer will merge the PR using squash-merge
   or rebase-merge depending on the commit history.

### PR Checklist

Before submitting, verify:

- [ ] Code compiles without errors on all target platforms (or at least your
      development OS).
- [ ] All existing tests pass.
- [ ] New code has corresponding tests.
- [ ] `gofmt -s -w .` has been run.
- [ ] `go vet ./...` reports no issues.
- [ ] Commit messages follow the conventional commits format.
- [ ] The PR description explains the change and its motivation.

---

## Project Structure

```
VoidVPN/
|-- cmd/
|   |-- voidvpn/
|       |-- main.go              # Application entry point
|
|-- internal/                    # Private application packages
|   |-- cli/                     # Cobra command definitions
|   |   |-- root.go              # Root command and flag registration
|   |   |-- connect.go           # voidvpn connect
|   |   |-- disconnect.go        # voidvpn disconnect
|   |   |-- status.go            # voidvpn status
|   |   |-- servers.go           # voidvpn servers (list/add/remove)
|   |   |-- keygen.go            # voidvpn keygen
|   |   |-- config.go            # voidvpn config
|   |   |-- version.go           # voidvpn version
|   |
|   |-- config/                  # Configuration and server management
|   |   |-- config.go            # YAML config load/save
|   |   |-- paths.go             # XDG/platform config directories
|   |   |-- server.go            # Server CRUD operations
|   |   |-- import.go            # WireGuard .conf file import
|   |
|   |-- daemon/                  # Background VPN daemon
|   |   |-- daemon.go            # Daemon lifecycle management
|   |   |-- ipc.go               # IPC server/client (Unix/Windows)
|   |   |-- ipc_unix.go          # Unix domain socket transport
|   |   |-- ipc_windows.go       # Named pipe transport
|   |   |-- state.go             # Connection state machine
|   |
|   |-- keystore/                # Cryptographic key storage
|   |   |-- keystore.go          # Keystore interface
|   |   |-- keyring.go           # OS keyring backend
|   |   |-- file.go              # File-based fallback backend
|   |
|   |-- network/                 # Network configuration
|   |   |-- interface.go         # Network interface management
|   |   |-- dns.go               # DNS configuration interface
|   |   |-- dns_windows.go       # DNS via netsh
|   |   |-- dns_linux.go         # DNS via resolv.conf
|   |   |-- routes.go            # Route table management interface
|   |   |-- routes_windows.go    # Routes via netsh/route
|   |   |-- routes_linux.go      # Routes via ip route
|   |
|   |-- platform/                # Platform abstraction
|   |   |-- admin.go             # Privilege escalation interface
|   |   |-- admin_windows.go     # Windows UAC elevation
|   |   |-- admin_unix.go        # Unix sudo/root check
|   |   |-- tun.go               # TUN device interface
|   |   |-- tun_windows.go       # Wintun backend
|   |   |-- tun_unix.go          # /dev/net/tun backend
|   |
|   |-- ui/                      # Terminal UI components
|   |   |-- styles.go            # Lipgloss style definitions
|   |   |-- banner.go            # ASCII art banner
|   |   |-- spinner.go           # Progress spinner
|   |   |-- status.go            # Connection status display
|   |   |-- table.go             # Table rendering
|   |
|   |-- wireguard/               # WireGuard protocol
|       |-- keys.go              # Keypair generation (Curve25519)
|       |-- config.go            # WireGuard config rendering
|       |-- tunnel.go            # Tunnel bring-up/tear-down
|       |-- device.go            # WireGuard device management
|       |-- ipc.go               # WireGuard userspace IPC (UAPI)
|
|-- pkg/                         # Public reusable packages
|   |-- version/
|       |-- version.go           # Build version info
|
|-- .github/
|   |-- workflows/
|       |-- ci.yml               # GitHub Actions CI/CD pipeline
|
|-- Dockerfile                   # Multi-stage Docker build
|-- Makefile                     # Build, test, lint targets
|-- go.mod                       # Go module definition
|-- go.sum                       # Dependency checksums
|-- .gitignore                   # Git ignore rules
```

### Package Roles

- **cmd/voidvpn**: Minimal entry point. Calls `cli.Execute()` and exits.
- **internal/cli**: All user-facing commands. Depends on every other internal
  package.
- **internal/config**: Reads/writes the YAML configuration file and manages
  server entries. Handles WireGuard `.conf` imports.
- **internal/daemon**: Runs the VPN connection in the background. Communicates
  with the CLI via IPC (Unix sockets on Linux/macOS, named pipes on Windows).
- **internal/keystore**: Stores WireGuard private keys. Uses the OS keyring
  (via `go-keyring`) with a file-based fallback.
- **internal/network**: Manages DNS settings and routing tables.
  Platform-specific implementations behind shared interfaces.
- **internal/platform**: Abstracts privilege elevation and TUN device creation
  across operating systems.
- **internal/ui**: Terminal UI components built with Charmbracelet (lipgloss,
  bubbles, bubbletea).
- **internal/wireguard**: Core WireGuard operations -- key generation, config
  rendering, tunnel lifecycle, device management via userspace IPC.
- **pkg/version**: Exposes build version, commit hash, and build date. Set via
  ldflags at compile time.

---

## Questions

If you have questions about contributing, open a GitHub issue with the `question`
label or start a discussion in the repository's Discussions tab.
