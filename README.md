# VoidVPN

**A fast, secure WireGuard VPN client for the command line.**

VoidVPN is a single-binary CLI tool that manages WireGuard VPN tunnels without requiring any external tools. It embeds the WireGuard protocol directly via `wireguard-go`, provides a styled terminal interface powered by Charmbracelet, and stores private keys securely in your operating system's native keyring.

---

## Features

- **Embedded WireGuard** -- Uses `wireguard-go` directly. No `wg`, `wg-quick`, or `wireguard-tools` required.
- **Single binary** -- One executable, zero runtime dependencies.
- **Cross-platform** -- Windows (primary), Linux, and macOS.
- **Styled terminal UI** -- Branded output with Charmbracelet (lipgloss, bubbletea, bubbles). Purple/cyan dark theme.
- **OS keyring integration** -- Private keys stored in Windows Credential Manager, macOS Keychain, or Linux secret-service.
- **Encrypted file fallback** -- Secure key storage even without a keyring provider.
- **WireGuard .conf import** -- Import existing WireGuard configuration files directly.
- **Kill switch** -- Optional traffic blocking if the VPN connection drops.
- **Daemon mode** -- Run the tunnel in the background.
- **Shell completions** -- Bash, Zsh, Fish, and PowerShell.

---

## Quick Start

### Install from source

```bash
go install github.com/voidvpn/voidvpn/cmd/voidvpn@latest
```

### Import an existing WireGuard config

```bash
voidvpn servers import /path/to/wg0.conf
```

### Connect

```bash
# Connect to a specific server (requires admin/root privileges)
sudo voidvpn connect myserver

# Or set a default and just run connect
voidvpn config set default_server myserver
sudo voidvpn connect
```

### Check status

```bash
voidvpn status
voidvpn status --watch    # live-updating display
voidvpn status --json     # machine-readable output
```

### Disconnect

```bash
voidvpn disconnect
```

---

## CLI Commands

### Global Flags

| Flag | Description |
|------|-------------|
| `-v`, `--verbose` | Enable verbose output |

### Commands

| Command | Description |
|---------|-------------|
| `voidvpn connect [server]` | Connect to a VPN server. Uses the default server if none specified. |
| `voidvpn disconnect` | Disconnect the active VPN session. |
| `voidvpn status` | Show current connection status. |
| `voidvpn servers list` | List all configured servers. |
| `voidvpn servers add <name>` | Add a new server configuration. |
| `voidvpn servers remove <name>` | Remove a server configuration. |
| `voidvpn servers import <file>` | Import a WireGuard `.conf` file. |
| `voidvpn keygen` | Generate a WireGuard keypair. |
| `voidvpn config show` | Display current configuration. |
| `voidvpn config set <key> <value>` | Set a configuration value. |
| `voidvpn version` | Show version and build information. |

### Command Flags

**connect**

| Flag | Description |
|------|-------------|
| `--daemon` | Run the tunnel in the background |

**status**

| Flag | Description |
|------|-------------|
| `--watch` | Live-updating status display (refreshes every 2s) |
| `--json` | Output status as JSON |

**servers add**

| Flag | Description |
|------|-------------|
| `--endpoint` | Server endpoint as `host:port` (required) |
| `--public-key` | Peer's WireGuard public key (required) |
| `--address` | Tunnel interface IP, e.g. `10.0.0.2/24` (required) |
| `--dns` | DNS servers, comma-separated |

**keygen**

| Flag | Description |
|------|-------------|
| `--save` | Store the private key in the OS keyring |
| `--name` | Name for the stored key (default: `default`) |

---

## Configuration

VoidVPN uses file-based configuration stored in YAML. Configuration files are placed in the standard location for each operating system:

| OS | Path |
|----|------|
| Windows | `%APPDATA%\VoidVPN\` |
| Linux | `~/.config/voidvpn/` (respects `$XDG_CONFIG_HOME`) |
| macOS | `~/Library/Application Support/voidvpn/` |

### Directory Layout

```
<config-dir>/
  config.yaml          # Application settings
  servers/             # One YAML file per server
  state/
    connection.json    # Runtime connection state
```

### Available Settings

| Key | Type | Description |
|-----|------|-------------|
| `log_level` | string | Logging verbosity: `debug`, `info`, `warn`, `error` |
| `default_server` | string | Server name used when `connect` is called without an argument |
| `auto_connect` | bool | Automatically connect on startup in daemon mode |
| `kill_switch` | bool | Block all traffic if the VPN connection drops |
| `dns_fallback` | list | Fallback DNS servers if the server-provided DNS fails |

### Server Configuration

Each server is stored as a separate YAML file in the `servers/` directory:

```yaml
name: myserver
endpoint: vpn.example.com:51820
public_key: <peer-public-key>
allowed_ips:
  - 0.0.0.0/0
  - ::/0
dns:
  - 1.1.1.1
  - 1.0.0.1
address: 10.0.0.2/24
persistent_keepalive: 25
mtu: 1420
```

---

## Building from Source

### Prerequisites

- **Go 1.22** or later
- **Git** (for version embedding)
- On Windows: `wintun.dll` must be present alongside the built binary

### Build

```bash
git clone https://github.com/voidvpn/voidvpn.git
cd voidvpn
make build
```

The binary is written to `./voidvpn` (or `voidvpn.exe` on Windows).

### Cross-compile

```bash
make build-windows   # Windows amd64
make build-linux     # Linux amd64
make build-darwin    # macOS amd64
make build-all       # All three platforms
```

### Other Makefile targets

| Target | Description |
|--------|-------------|
| `make test` | Run all tests |
| `make test-coverage` | Run tests with coverage report |
| `make vet` | Run `go vet` |
| `make fmt` | Format all source with `gofmt` |
| `make lint` | Run `golangci-lint` |
| `make install` | Install the binary to `$GOPATH/bin` |
| `make completions` | Generate shell completion scripts |
| `make clean` | Remove build artifacts |

---

## Docker

### Build the image

```bash
make docker-build
```

Or manually:

```bash
docker build -t voidvpn:latest .
```

### Run

The container requires `NET_ADMIN` capabilities and access to `/dev/net/tun` for tunnel creation:

```bash
docker run --rm -it \
  --cap-add=NET_ADMIN \
  --device=/dev/net/tun \
  -v /path/to/config:/home/voidvpn/.config/voidvpn \
  voidvpn:latest connect myserver
```

The image uses a multi-stage build (Go 1.22 Alpine builder, Alpine 3.19 runtime) and runs as a non-root user by default. The runtime image includes `iptables` and `iproute2` for route and DNS management.

---

## Security

### Key Storage

Private keys are stored in your operating system's native credential store:

| OS | Backend |
|----|---------|
| Windows | Windows Credential Manager |
| macOS | Keychain |
| Linux | Secret Service (via D-Bus) |

If no keyring provider is available, VoidVPN falls back to an encrypted file stored in the configuration directory.

### Privilege Requirements

Creating TUN interfaces and modifying system routes and DNS requires elevated privileges:

- **Windows** -- Run as Administrator
- **Linux/macOS** -- Run with `sudo` or as root

VoidVPN checks for sufficient privileges at startup and exits with a clear error if they are not present.

### Network Safety

- DNS and routes are restored on disconnect, crash, or signal interruption (`SIGINT`/`SIGTERM`) via deferred cleanup and signal handlers.
- The default routing strategy uses the `0.0.0.0/1` + `128.0.0.0/1` split to override the default route without deleting it, ensuring clean rollback.
- The optional kill switch blocks all non-tunnel traffic to prevent leaks if the connection drops unexpectedly.

---

## Project Structure

```
cmd/voidvpn/
  main.go                    # Entry point

internal/
  cli/                       # Cobra command definitions
    root.go                  # Root command, global flags
    connect.go               # connect command
    disconnect.go            # disconnect command
    status.go                # status command
    servers.go               # servers list/add/remove/import
    config.go                # config show/set
    keygen.go                # keygen command
    version.go               # version command

  wireguard/                 # WireGuard tunnel management
    tunnel.go                # Tunnel lifecycle
    device.go                # wireguard-go device wrapper
    config.go                # WG config types
    keys.go                  # Key generation and parsing
    ipc.go                   # IPC config string builder

  network/                   # Platform-specific networking
    dns.go                   # DNS manager interface
    dns_windows.go           # Windows DNS (netsh)
    dns_linux.go             # Linux DNS (resolv.conf)
    routes.go                # Route manager interface
    routes_windows.go        # Windows routes
    routes_linux.go          # Linux routes
    interface.go             # Interface address utilities

  config/                    # Application configuration
    config.go                # Config struct, Load/Save
    paths.go                 # OS-specific config paths
    server.go                # Server config CRUD
    import.go                # WireGuard .conf importer

  keystore/                  # Secure key storage
    keystore.go              # Keystore interface
    keyring.go               # OS keyring backend
    file.go                  # Encrypted file fallback

  daemon/                    # Background process and IPC
    daemon.go                # Daemon lifecycle
    ipc.go                   # IPC protocol
    ipc_windows.go           # Windows named pipes
    ipc_unix.go              # Unix domain sockets
    state.go                 # Connection state file

  ui/                        # Terminal UI (Charmbracelet)
    styles.go                # Brand colors and lipgloss styles
    spinner.go               # Connection spinner
    table.go                 # Server list table
    status.go                # Status display formatting
    banner.go                # ASCII art banner

  platform/                  # Platform abstraction
    admin.go                 # Privilege detection interface
    admin_windows.go         # Windows elevation check
    admin_unix.go            # Unix root check
    tun.go                   # TUN device interface
    tun_windows.go           # Wintun integration
    tun_unix.go              # Linux/macOS TUN

pkg/version/
  version.go                 # Version info (set via ldflags)
```

---

## Contributing

Contributions are welcome. Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines on submitting issues and pull requests.

---

## License

VoidVPN is released under the [MIT License](LICENSE).
