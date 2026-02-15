# VoidVPN Architecture Guide

## Overview

VoidVPN is a CLI-based WireGuard VPN client written in Go. It compiles to a single
binary with no external runtime dependencies -- there is no need to install `wg`,
`wg-quick`, or `wireguard-tools`. The WireGuard protocol implementation is embedded
directly into the binary via the `wireguard-go` library, and TUN device creation on
Windows is handled through the bundled `wintun.dll`.

### Design philosophy

1. **Single binary** -- Ship one executable that contains everything needed to
   establish, manage, and tear down WireGuard tunnels.
2. **No external dependencies** -- The binary embeds wireguard-go and drives TUN
   devices, DNS, and routes itself.  Users never invoke `wg` or `wg-quick`.
3. **Windows-first, cross-platform** -- Primary development target is Windows, with
   full Linux and macOS support via Go build tags.
4. **Secure by default** -- Private keys live in the OS keyring (Windows Credential
   Manager, macOS Keychain, Linux secret-service) with an AES-256-GCM encrypted file
   fallback.  Config directories are created with 0700 permissions.
5. **Terminal-native UX** -- Styled output via Charmbracelet (lipgloss, bubbletea,
   bubbles) with branded purple/cyan colors.

### Technology stack

| Layer | Technology |
|-------|-----------|
| Language | Go 1.22+ |
| CLI framework | Cobra + Viper |
| WireGuard protocol | wireguard-go (golang.zx2c4.com/wireguard) |
| TUN driver (Windows) | Wintun (golang.zx2c4.com/wintun) |
| Terminal UI | Charmbracelet lipgloss, bubbletea, bubbles |
| Secure storage | zalando/go-keyring |
| Config format | YAML (gopkg.in/yaml.v3) |
| WireGuard config import | INI (gopkg.in/ini.v1) |

---

## Connection Flow

The following describes the sequence of operations that occur when a user runs
`voidvpn connect <server>`.

### 1. Pre-flight checks

- Verify the process has administrator/root privileges (`platform.IsAdmin()`).
  On Windows this checks membership in the BUILTIN\Administrators SID.
  On Unix it checks `os.Geteuid() == 0`.
- Verify no existing connection is active by checking for the connection state
  file at `<config-dir>/state/connection.json`.
- Resolve the server name: use the argument if provided, otherwise fall back to
  the `default_server` value from `config.yaml`.

### 2. Load credentials

- Load the server configuration from `<config-dir>/servers/<name>.yaml`.
- Load the WireGuard private key from the keystore.  The keystore first tries
  a key stored under the server name, then falls back to a key named "default".

### 3. Create TUN device

`platform.CreateTUN("voidvpn0", mtu)` is called.  The implementation varies by
platform:

- **Windows:** Calls `tun.CreateTUN()` from wireguard-go which loads `wintun.dll`
  and creates a Wintun adapter.
- **Linux:** Opens `/dev/net/tun` and creates a TUN interface.
- **macOS:** Creates a utun device.

The default MTU is 1420.

### 4. Create and configure WireGuard device

- `device.NewDevice(tunDev, bind, logger)` creates the wireguard-go device using
  the TUN device and a default UDP bind.
- `device.IpcSet(ipcConfig)` configures the device with the WireGuard settings:
  private key (hex-encoded), peer public key, endpoint, allowed IPs, optional
  preshared key, and persistent keepalive interval.
- `device.Up()` activates the device and begins the WireGuard handshake.

### 5. Assign IP address

`network.AssignAddress(interfaceName, address)` assigns the tunnel IP to the
interface:

- **Windows:** `netsh interface ip set address name=<iface> static <ip> <mask>`
- **Linux:** `ip addr add <prefix> dev <iface>` followed by `ip link set <iface> up`

### 6. Configure DNS

The DNS manager saves the current DNS settings before overwriting them:

- **Windows:** Saves existing DNS via `netsh interface ip show dns`, then sets
  new servers with `netsh interface ip set dns` and `netsh interface ip add dns`.
- **Linux:** Backs up `/etc/resolv.conf` in memory, then writes a new file with
  `nameserver` entries for each configured DNS server.

DNS server addresses are validated as valid IPs before being applied.

### 7. Configure routes

The route manager installs three routes:

1. **Endpoint route:** The VPN server's IP is routed via the current default
   gateway so the encrypted WireGuard UDP traffic continues to flow over the
   physical network.
2. **0.0.0.0/1** via the tunnel gateway.
3. **128.0.0.0/1** via the tunnel gateway.

These two routes together cover the entire IPv4 address space and override the
default route (0.0.0.0/0) because they are more specific, without deleting the
original default route.  See "Route Strategy" below for details.

- **Windows:** Uses `route add <network> mask <mask> <gateway> metric 5`.
- **Linux:** Uses `ip route add <cidr> via <gateway>` or `ip route add <cidr> dev <iface>`.

### 8. Start IPC server

The daemon starts an IPC server for `disconnect` and `status` commands:

- **Windows:** TCP listener on `127.0.0.1:41820` with token-based authentication.
- **Unix:** Unix domain socket at `$XDG_RUNTIME_DIR/voidvpn.sock` (or a
  user-specific temporary directory as fallback), created with umask 0077.

### 9. Write connection state

A JSON state file is written to `<config-dir>/state/connection.json` containing
the server name, connection timestamp, interface name, tunnel IP, endpoint, PID,
and traffic counters.

### 10. Block on signal

The daemon enters a blocking select waiting for either:

- An OS signal (SIGINT or SIGTERM).
- Context cancellation (triggered by an IPC `disconnect` command).

### 11. Cleanup (disconnect)

Cleanup runs in reverse order via `defer`:

1. Close the IPC server (remove token file or socket).
2. Remove VPN routes (reverse order of addition).
3. Restore DNS settings (DHCP on Windows, original resolv.conf on Linux).
4. `device.Close()` to bring down the WireGuard device and close the TUN.
5. Remove the connection state file.

---

## Package Architecture

### cmd/voidvpn

Entry point.  `main.go` calls `cli.Execute()` and exits with code 1 on error.

### internal/cli

Cobra command definitions.  Each file defines one command or command group:

| File | Command | Description |
|------|---------|-------------|
| root.go | `voidvpn` | Root command, `--verbose` flag, ensures config dirs exist |
| connect.go | `voidvpn connect [server]` | Privilege check, load config/key, create tunnel, run daemon |
| disconnect.go | `voidvpn disconnect` | Send `disconnect` via IPC |
| status.go | `voidvpn status` | Send `status` via IPC or read state file; `--watch`, `--json` |
| servers.go | `voidvpn servers {list,add,remove,import}` | Server CRUD and .conf import |
| keygen.go | `voidvpn keygen` | Generate WireGuard keypair; `--save` to persist in keystore |
| config.go | `voidvpn config {show,set}` | Read/write app configuration |
| version.go | `voidvpn version` | Print version, commit, build date, OS/arch |

### internal/wireguard

Core WireGuard tunnel management:

- **tunnel.go** -- `Tunnel` struct that orchestrates TUN creation, device setup,
  connect/disconnect lifecycle, and status queries.
- **device.go** -- Wrapper around `device.Device` from wireguard-go.  Handles
  configuration via IPC strings, up/down lifecycle, traffic stats retrieval,
  and CIDR address parsing.
- **config.go** -- `TunnelConfig` struct holding all WireGuard parameters
  (private key, address, DNS, MTU, peer settings).
- **keys.go** -- Key generation using `crypto/rand` and Curve25519 scalar
  base multiplication.  Produces base64-encoded keypairs.
- **ipc.go** -- Builds the IPC configuration string for `device.IpcSet()`.
  Converts base64 keys to hex, validates inputs against newline injection.

### internal/network

Platform-specific network configuration:

- **dns.go** -- `DNSManager` interface with `Set()` and `Restore()` methods.
- **dns_windows.go** -- Windows implementation using `netsh` commands.
- **dns_linux.go** -- Linux implementation that overwrites `/etc/resolv.conf`.
- **routes.go** -- `RouteManager` interface with `AddVPNRoutes()` and
  `RemoveVPNRoutes()` methods.
- **routes_windows.go** -- Windows implementation using the `route` command.
- **routes_linux.go** -- Linux implementation using `ip route`.
- **interface.go** -- Cross-platform utilities: `AssignAddress()`,
  `ExtractGateway()`, `ExtractEndpointHost()`, `prefixToMask()`.

### internal/config

Application configuration:

- **config.go** -- `AppConfig` struct with `Load()`, `Save()`, `Get()`, `Set()`
  methods.  Reads/writes `config.yaml`.
- **paths.go** -- OS-specific directory resolution (`ConfigDir()`, `ServersDir()`,
  `StateDir()`, `ConfigFile()`, `StateFile()`, `EnsureDirs()`).
- **server.go** -- `ServerConfig` struct with CRUD operations (`LoadServer`,
  `SaveServer`, `RemoveServer`, `ListServers`).  Server names are validated
  against a regex and checked for path traversal.
- **import.go** -- Parses standard WireGuard `.conf` files (INI format) and
  converts them to `ServerConfig` structs.  Extracts private key separately.

### internal/keystore

Secure private key storage:

- **keystore.go** -- `Keystore` interface (`Store`, `Load`, `Delete`, `Exists`).
  `New()` factory probes the OS keyring and falls back to encrypted file storage
  if the keyring is unavailable.
- **keyring.go** -- OS keyring backend using `zalando/go-keyring`.  Stores keys
  under the service name "VoidVPN".
- **file.go** -- AES-256-GCM encrypted file backend.  Keys are stored as base64
  in `<config-dir>/keys/<name>.key`.  Encryption key is derived from
  SHA-256(random-salt || hostname || home-directory).  The salt is generated once
  and saved as `<config-dir>/keys/.salt`.  Key names are validated against a
  strict regex and checked for path traversal.

### internal/daemon

Background process and IPC:

- **daemon.go** -- `Daemon` struct that orchestrates the full connection lifecycle:
  connect tunnel, assign IP, set DNS, add routes, save state, start IPC, wait
  for signal, cleanup.
- **state.go** -- `ConnectionState` JSON serialization.  `SaveState()`,
  `LoadState()`, `ClearState()`, `IsConnected()`.
- **ipc.go** -- `IPCRequest`/`IPCResponse` types and JSON marshal/unmarshal
  helpers.  Commands: `"status"` and `"disconnect"`.
- **ipc_windows.go** -- TCP IPC server on `127.0.0.1:41820` with 32-byte hex
  token authentication.  Token is stored in `<config-dir>/state/ipc.token`.
- **ipc_unix.go** -- Unix domain socket IPC server.  Socket created with umask
  0077 for owner-only access.  Located at `$XDG_RUNTIME_DIR/voidvpn.sock`.

### internal/ui

Terminal UI components using Charmbracelet:

- **styles.go** -- Brand colors (`#7B2FBE` purple, `#00D4FF` cyan) and lipgloss
  styles for titles, labels, success/warning/error messages, and dim text.
- **banner.go** -- ASCII art "VoidVPN" logo rendered in brand colors.
- **spinner.go** -- Bubbletea spinner model shown during connection.
- **table.go** -- Styled table renderer for server lists.
- **status.go** -- Connection status display formatting.

### internal/platform

Platform-specific utilities:

- **admin.go** -- `IsAdmin()` public function dispatching to platform implementation.
- **admin_windows.go** -- Checks Windows BUILTIN\Administrators SID membership.
- **admin_unix.go** -- Checks `os.Geteuid() == 0`.
- **tun.go** -- `CreateTUN()` public function dispatching to platform implementation.
- **tun_windows.go** -- Wintun-based TUN creation, plus `GetInterfaceLUID()`.
- **tun_unix.go** -- Standard wireguard-go TUN creation for Linux/macOS.

### pkg/version

Public version information:

- **version.go** -- `Version`, `Commit`, `BuildDate` variables set via `-ldflags`
  at build time.  `Full()` returns a formatted string with OS/arch.  `Short()`
  returns just the version tag.

---

## Platform Abstraction

VoidVPN uses Go build tags to provide platform-specific implementations behind
common interfaces.  The pattern is consistent: a public function in a shared file
dispatches to an unexported function that has platform-specific implementations.

### Build tag files

| Shared interface | Windows (`//go:build windows`) | Unix (`//go:build !windows`) |
|------------------|-------------------------------|------------------------------|
| platform/tun.go | platform/tun_windows.go | platform/tun_unix.go |
| platform/admin.go | platform/admin_windows.go | platform/admin_unix.go |
| network/dns.go | network/dns_windows.go | network/dns_linux.go |
| network/routes.go | network/routes_windows.go | network/routes_linux.go |
| daemon/ipc.go | daemon/ipc_windows.go | daemon/ipc_unix.go |

### How it works

Each shared file defines an interface or public function:

    // dns.go
    type DNSManager interface {
        Set(iface string, servers []string) error
        Restore() error
    }

    func NewDNSManager() DNSManager {
        return newDNSManager()   // dispatches to build-tag file
    }

The platform files define `newDNSManager()` with the appropriate build tag.
On Windows this returns a `windowsDNS` struct that uses `netsh`; on Linux it
returns a `unixDNS` struct that writes `/etc/resolv.conf`.

The same pattern applies to TUN creation (`createTUN`), admin checks (`isAdmin`),
route management (`newRouteManager`), and IPC (`NewIPCServer`, `SendIPCRequest`).

### Network interface configuration

`network/interface.go` uses `runtime.GOOS` rather than build tags for IP address
assignment, since both Windows (`netsh`) and Linux (`ip addr`) implementations
are small enough to coexist in a single file.

---

## IPC Protocol

VoidVPN uses a simple JSON-based request/response protocol over a stream
connection.  Messages are newline-delimited.

### Transport

- **Windows:** TCP on `127.0.0.1:41820`.
- **Unix:** Unix domain socket at `$XDG_RUNTIME_DIR/voidvpn.sock` (or
  `/tmp/voidvpn-<uid>/voidvpn.sock`).

### Authentication (Windows only)

Because TCP sockets on localhost can be connected to by any local user, Windows
uses token-based authentication:

1. When the IPC server starts, it generates a 32-byte random token via
   `crypto/rand` and hex-encodes it (64 characters).
2. The token is written to `%APPDATA%\VoidVPN\state\ipc.token` with mode 0600.
3. Clients must send the token as the first line before sending any command.
4. The server validates the token and rejects connections with a mismatched token.
5. On shutdown, the token file is deleted.

Unix sockets are inherently restricted by filesystem permissions (umask 0077),
so no token authentication is needed.

### Request format

    {"command": "<command-name>"}

Supported commands: `"status"`, `"disconnect"`.

### Response format

    {"success": true, "state": { ... }}       // success with optional state
    {"success": false, "error": "message"}    // failure

The `state` field is included only for `status` responses and contains the full
`ConnectionState` object (server name, connected_at, interface_name, tunnel_ip,
endpoint, pid, tx_bytes, rx_bytes).

### Timeouts

- Connection timeout: 3 seconds.
- Per-connection deadline: 5 seconds (covers auth + request + response).

---

## Security Model

### Private key storage

Private keys are never stored in plain text on disk.

**Primary: OS keyring**

The `keyringStore` uses `zalando/go-keyring` to store keys in the
platform-native credential store:

- Windows: Windows Credential Manager
- macOS: Keychain
- Linux: D-Bus secret-service (GNOME Keyring, KDE Wallet)

The factory function `keystore.New()` probes the keyring with a test write.
If it fails (e.g., no D-Bus session, headless server), it falls back to the
file store.

**Fallback: Encrypted files**

Keys are encrypted with AES-256-GCM and stored as base64 in
`<config-dir>/keys/<name>.key`.  The encryption key is derived as follows:

1. A 32-byte random salt is generated once and stored at `<config-dir>/keys/.salt`.
2. A material string is formed: `"voidvpn:<hostname>:<home-directory>"`.
3. The encryption key is `SHA-256(salt || material)`.

This is not a password-based scheme -- it is machine-bound.  The intent is to
prevent casual exfiltration of key files; it does not protect against an attacker
with full access to the machine.  For stronger protection, use the OS keyring.

GCM nonces are randomly generated per encryption operation.

### Input validation

- **Server names:** Validated against `^[a-zA-Z0-9][a-zA-Z0-9 _-]{0,62}$`.
  After validation, names are lowercased and spaces replaced with hyphens for
  filesystem use.  A path traversal check ensures the resolved path stays under
  the servers directory.
- **Key names:** Validated against `^[a-zA-Z0-9][a-zA-Z0-9_-]{0,62}$` with the
  same path traversal check.
- **DNS server addresses:** Validated as IP addresses via `net.ParseIP()` before
  being passed to system commands.
- **Route addresses:** Validated via `net.ParseIP()` before being passed to
  `route add` or `ip route add`.
- **Interface names (Linux):** Validated against `^[a-zA-Z0-9_-]+$`.
- **WireGuard IPC strings:** Endpoint and allowed-IP values are checked for
  newline characters to prevent IPC injection.

### Privilege model

TUN device creation and network configuration require elevated privileges:

- **Windows:** The process must be running as Administrator (checked via
  BUILTIN\Administrators SID membership).
- **Unix:** The process must be running as root (euid == 0).

The `connect` command checks privileges before attempting any operations and
provides a clear error message if not elevated.

### File permissions

- Config directories: created with 0700 (owner-only).
- Config files: written with 0600 (owner read/write only).
- IPC token file: written with 0600.
- Unix socket: created with umask 0077 (owner-only).

---

## Route Strategy

VoidVPN uses the "split default route" technique to route all traffic through the
VPN tunnel without modifying the existing default route.

### The problem

A naive approach would replace the default route (0.0.0.0/0) with one pointing
at the tunnel.  This has two issues:

1. The encrypted WireGuard UDP traffic itself needs to reach the real server
   endpoint over the physical network.  Routing it through the tunnel creates
   a loop.
2. Deleting and restoring the default route is fragile.  If the VPN process
   crashes, the default route is lost and the host loses connectivity.

### The solution

Three routes are added:

1. **Endpoint route:** `<server-ip>/32 via <original-default-gateway>` -- ensures
   WireGuard UDP packets reach the server over the physical link.
2. **0.0.0.0/1 via <tunnel-gateway>** -- covers addresses 0.0.0.0 through
   127.255.255.255.
3. **128.0.0.0/1 via <tunnel-gateway>** -- covers addresses 128.0.0.0 through
   255.255.255.255.

Together, routes 2 and 3 cover the entire IPv4 address space.  Because /1 is
more specific than /0, they take priority over the existing default route.  The
original default route remains in the routing table, untouched.  If the VPN
process crashes and fails to clean up, the /1 routes become stale and
eventually the original default route resumes handling traffic.

### Implementation

- **Windows:** `route add <network> mask <mask> <gateway> metric 5`
  (and `route delete` on teardown).
- **Linux:** `ip route add <cidr> dev <iface>` for tunnel routes,
  `ip route add <endpoint>/32 via <default-gw>` for the endpoint route
  (and `ip route delete` on teardown).

Routes are removed in reverse order of addition during cleanup.

---

## Key Design Decisions

### Why wireguard-go is embedded

Embedding wireguard-go directly means VoidVPN has no runtime dependency on
`wireguard-tools`, `wg-quick`, or a kernel WireGuard module.  This makes
installation trivial (single binary + wintun.dll on Windows) and avoids version
mismatches between the client and system WireGuard tools.  The Go
implementation of WireGuard is well-tested and used by the official WireGuard
apps on Android, iOS, and macOS.

### Why Cobra + Viper

Cobra is the de facto standard CLI framework in the Go ecosystem.  It provides
subcommand routing, flag parsing, shell completions, and usage generation out
of the box.  Viper is its companion for configuration management.  Together
they reduce boilerplate and provide a familiar developer experience.

### Why Charmbracelet

The Charmbracelet ecosystem (lipgloss, bubbletea, bubbles) provides modern,
composable terminal UI primitives.  lipgloss handles styled output without
ANSI escape code management.  bubbletea provides an Elm-like architecture for
interactive components (spinners, tables, live-updating status).  These
libraries work correctly across Windows Terminal, PowerShell, CMD, and Unix
terminals.

### Why TCP IPC on Windows

Windows does not support Unix domain sockets in all configurations (they
require Windows 10 1803+ and are not available in all Go standard library
versions).  Named pipes are an option but add complexity.  A TCP listener on
localhost with token-based authentication is simple, portable, and easy to
debug.  The auth token prevents other local users from controlling the VPN.

### Why YAML for configuration

YAML is human-readable and supports comments, making it suitable for
configuration files that users may edit by hand.  Server configs are stored as
separate YAML files (one per server) for easy management and to avoid merge
conflicts when multiple servers are configured.

### Why OS keyring with encrypted file fallback

The OS keyring is the most secure option for key storage on desktop systems.
However, it is not always available (e.g., headless Linux servers, SSH
sessions, containers).  The encrypted file fallback ensures VoidVPN works
everywhere, with the tradeoff that the encryption is machine-bound rather
than password-bound.

---

## Configuration File Locations

| Platform | Config directory | Config file | Server configs | State file |
|----------|-----------------|-------------|----------------|------------|
| Windows | `%APPDATA%\VoidVPN\` | `config.yaml` | `servers\*.yaml` | `state\connection.json` |
| Linux | `~/.config/voidvpn/` | `config.yaml` | `servers/*.yaml` | `state/connection.json` |
| macOS | `~/Library/Application Support/voidvpn/` | `config.yaml` | `servers/*.yaml` | `state/connection.json` |

Linux respects `$XDG_CONFIG_HOME` if set.

---

## Build System

The project uses a Makefile with the following key targets:

- `make build` -- Build for the current platform.
- `make build-windows` / `make build-linux` / `make build-darwin` -- Cross-compile.
- `make build-all` -- Build for all three platforms.
- `make test` -- Run all tests with verbose output.
- `make test-coverage` -- Run tests with coverage reporting.
- `make install` -- Install to `$GOPATH/bin`.
- `make docker-build` -- Build Docker image.
- `make completions` -- Generate shell completion scripts (bash, zsh, fish, PowerShell).

Version information is injected via `-ldflags`:

    -X github.com/voidvpn/voidvpn/pkg/version.Version=<tag>
    -X github.com/voidvpn/voidvpn/pkg/version.Commit=<short-hash>
    -X github.com/voidvpn/voidvpn/pkg/version.BuildDate=<ISO8601>
