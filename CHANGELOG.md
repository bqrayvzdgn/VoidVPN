# Changelog

All notable changes to VoidVPN will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- WireGuard tunnel management with connect, disconnect, and status commands.
- Server configuration CRUD operations (add, remove, list, show) with YAML
  persistence.
- WireGuard `.conf` file import for migrating existing server configurations.
- WireGuard keypair generation using Curve25519 via `golang.zx2c4.com/wireguard`.
- OS keyring integration for secure private key storage, with file-based
  encrypted fallback for headless environments.
- Styled terminal UI built with Charmbracelet (lipgloss, bubbles, bubbletea),
  including ASCII banner, progress spinners, status displays, and table
  rendering.
- Cross-platform support for Windows, Linux, and macOS with platform-specific
  implementations for TUN devices, privilege elevation, DNS, and routing.
- Split routing using `0.0.0.0/1` and `128.0.0.0/1` to route all traffic
  through the VPN tunnel without replacing the default gateway.
- DNS management via `netsh` on Windows and `/etc/resolv.conf` manipulation
  on Linux, with automatic restoration on disconnect.
- IPC daemon for persistent background VPN connections, using Unix domain
  sockets on Linux/macOS and named pipes on Windows.
- Docker support with a multi-stage Alpine-based build producing a minimal
  runtime image.
- CI/CD pipeline with GitHub Actions: tests on Ubuntu, Windows, and macOS;
  cross-compilation for amd64 and arm64; automated release artifact uploads
  on version tags.

### Security

- Token-based IPC authentication between the CLI client and the background
  daemon to prevent unauthorized local processes from controlling VPN
  connections.
- Input validation on all system commands (DNS, routing, interface management)
  to prevent command injection via untrusted configuration values.
- Path traversal prevention on configuration file import and server name
  handling to block writes outside the designated configuration directory.
