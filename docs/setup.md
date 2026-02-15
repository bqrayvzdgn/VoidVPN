# VoidVPN Setup and Usage Guide

## Prerequisites

- **Go 1.22 or later** -- Required for building from source.
- **Git** -- Required for cloning the repository and for version stamping during
  builds.
- **Administrator/root privileges** -- Required for VPN operations (TUN device
  creation, DNS changes, route manipulation).  Not required for configuration
  management or key generation.
- **Windows:** `wintun.dll` must be present alongside the `voidvpn.exe` binary.
  It is included in release archives.
- **Linux:** The `ip` command (iproute2) must be installed.  It is present on
  virtually all modern distributions.

---

## Installation

### From source

Clone the repository and build:

    git clone https://github.com/voidvpn/voidvpn.git
    cd voidvpn
    make build

This produces a `voidvpn` binary (or `voidvpn.exe` on Windows) in the project
root.  Move it to a directory in your PATH.

Alternatively, use `go install`:

    go install github.com/voidvpn/voidvpn/cmd/voidvpn@latest

Or install directly with `make`:

    make install

This places the binary in `$GOPATH/bin`.

### Cross-compilation

Build for a specific platform:

    make build-windows    # produces voidvpn.exe
    make build-linux      # produces voidvpn-linux
    make build-darwin     # produces voidvpn-darwin

Build for all platforms:

    make build-all

### From releases

Download the appropriate archive for your platform from the GitHub releases page.
Extract it and place the binary in your PATH.  On Windows, ensure `wintun.dll` is
in the same directory as `voidvpn.exe`.

### Verify installation

    voidvpn version

This prints the version, commit hash, build date, and platform.

---

## First Steps

### 1. Generate a WireGuard keypair

    voidvpn keygen --save

This generates a Curve25519 keypair, prints both keys, and stores the private
key in the OS keyring under the name "default".  The public key should be
provided to your WireGuard server administrator.

To save the key under a specific name (matching a server name):

    voidvpn keygen --save --name myserver

### 2. Add a server

Option A -- Add manually with flags:

    voidvpn servers add myserver \
      --endpoint "vpn.example.com:51820" \
      --public-key "ServerPublicKeyBase64=" \
      --address "10.0.0.2/24" \
      --dns "1.1.1.1,1.0.0.1"

Option B -- Import from a standard WireGuard `.conf` file:

    voidvpn servers import /path/to/myserver.conf

The import command reads the `[Interface]` and `[Peer]` sections, extracts the
server name from the filename, and stores the private key in the keystore
automatically.

### 3. Connect

Run from an elevated terminal (Administrator on Windows, root on Linux):

    voidvpn connect myserver

VoidVPN will create a TUN device, establish the WireGuard tunnel, configure DNS,
add routes, and display a connection spinner.  Once connected, it blocks until
you disconnect or the process receives a termination signal.

### 4. Check status

In a separate terminal:

    voidvpn status

Add `--watch` for live-updating output, or `--json` for machine-readable output:

    voidvpn status --watch
    voidvpn status --json

### 5. Disconnect

    voidvpn disconnect

This sends a disconnect command via IPC to the running VPN process, which then
tears down the tunnel, restores DNS, removes routes, and exits cleanly.

---

## Configuration Reference

Application configuration is stored in `config.yaml` in the platform-specific
config directory (see below).  Values can be viewed and set via the CLI.

### View all configuration

    voidvpn config show

### Set a value

    voidvpn config set <key> <value>

### Configuration keys

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `log_level` | string | `"info"` | Log verbosity. One of: `debug`, `info`, `warn`, `error`. |
| `default_server` | string | `""` (empty) | Server name to use when `voidvpn connect` is called without an argument. |
| `auto_connect` | bool | `false` | Reserved for future use. When enabled, automatically connect on daemon startup. |
| `kill_switch` | bool | `false` | Reserved for future use. When enabled, block all traffic if the VPN connection drops. |
| `dns_fallback` | []string | `["1.1.1.1", "8.8.8.8"]` | Fallback DNS servers used when the server-specified DNS is unreachable. |

### Examples

    voidvpn config set log_level debug
    voidvpn config set default_server myserver
    voidvpn config set kill_switch true

---

## Server Configuration

### Adding servers manually

    voidvpn servers add <name> \
      --endpoint <host:port> \
      --public-key <base64-key> \
      --address <tunnel-ip/prefix> \
      [--dns <server1,server2>]

Required flags:

- `--endpoint` -- The WireGuard server endpoint in `host:port` format.
- `--public-key` -- The server's WireGuard public key (base64-encoded).
- `--address` -- The tunnel IP address assigned to this client (e.g., `10.0.0.2/24`).

Optional flags:

- `--dns` -- Comma-separated list of DNS servers to use when connected.
  Defaults to `1.1.1.1, 1.0.0.1`.

Default values applied automatically:

- `allowed_ips`: `["0.0.0.0/0", "::/0"]` (route all traffic)
- `persistent_keepalive`: 25 seconds
- `mtu`: 1420

### Importing WireGuard .conf files

    voidvpn servers import <path-to-file.conf>

The importer reads standard WireGuard configuration files in INI format.  It
extracts:

- From `[Interface]`: Address, DNS, MTU, PrivateKey.
- From `[Peer]`: PublicKey, Endpoint, AllowedIPs, PresharedKey, PersistentKeepalive.

The server name is derived from the filename (without extension).  The private
key, if present in the file, is automatically stored in the keystore under the
server name.

### Listing servers

    voidvpn servers list

Displays a styled table with columns: Name, Endpoint, Address, DNS.

### Removing servers

    voidvpn servers remove <name>

Removes the server configuration file.  This does not remove the associated
private key from the keystore.

### Server config file format

Server configs are stored as individual YAML files in the `servers/` subdirectory
of the config directory.  The filename is the lowercase, hyphenated server name
with a `.yaml` extension.

Example (`servers/myserver.yaml`):

    name: myserver
    endpoint: vpn.example.com:51820
    public_key: ServerPublicKeyBase64=
    allowed_ips:
      - 0.0.0.0/0
      - "::/0"
    dns:
      - 1.1.1.1
      - 1.0.0.1
    address: 10.0.0.2/24
    persistent_keepalive: 25
    mtu: 1420

---

## Connecting

### Basic connection

    voidvpn connect <server-name>

Requires administrator/root privileges.  If no server name is provided, the
`default_server` from `config.yaml` is used.

### Daemon mode

    voidvpn connect <server-name> --daemon

Runs the VPN connection in the background.  The process detaches from the
terminal and continues running until explicitly disconnected.

### Connection lifecycle

1. Pre-flight: privilege check, duplicate connection check, server config load.
2. Key load: retrieves private key from keystore (tries server name, then "default").
3. Tunnel: TUN device creation, WireGuard device setup, handshake.
4. Network: IP assignment, DNS configuration, route installation.
5. IPC: starts the IPC server for disconnect/status commands.
6. State: writes `connection.json` with connection metadata.
7. Wait: blocks on OS signal (Ctrl+C) or IPC disconnect command.
8. Cleanup: IPC close, route removal, DNS restore, device teardown, state file removal.

### Status monitoring

    voidvpn status              # one-time status
    voidvpn status --watch      # live-updating display
    voidvpn status --json       # JSON output for scripting

Status includes: server name, tunnel IP, endpoint, connection duration, and
transmit/receive byte counts.

---

## Troubleshooting

### "administrator/root privileges required"

VPN operations (connect/disconnect) require elevated privileges because they
create TUN devices and modify system DNS and routing tables.

- **Windows:** Right-click your terminal (Command Prompt, PowerShell, or Windows
  Terminal) and select "Run as administrator", or use `gsudo` if installed.
- **Linux:** Use `sudo voidvpn connect <server>`.

Key generation and configuration management do not require elevated privileges.

### DNS not restoring after disconnect

If DNS settings are not properly restored after a crash or forced termination:

- **Windows:** Open an elevated Command Prompt and run:

      netsh interface ip set dns name="Wi-Fi" dhcp
      netsh interface ip set dns name="Ethernet" dhcp

  Replace "Wi-Fi" or "Ethernet" with your active network adapter name.

- **Linux:** Restore your original `/etc/resolv.conf`.  If you use
  `systemd-resolved`, restart it:

      sudo systemctl restart systemd-resolved

  If you use NetworkManager:

      sudo systemctl restart NetworkManager

### "VPN is not running (could not connect to IPC)"

This error from `voidvpn disconnect` or `voidvpn status` means no VPN process
is running, or the IPC server is not reachable.

- Verify the VPN process is running: check for a `voidvpn` process in your
  task manager or with `ps aux | grep voidvpn`.
- If the process crashed, clean up the stale state file manually:
  - **Windows:** Delete `%APPDATA%\VoidVPN\state\connection.json`
  - **Linux:** Delete `~/.config/voidvpn/state/connection.json`

### "already connected"

VoidVPN checks for an existing connection state file before connecting.  If a
previous connection was not cleaned up properly:

1. Try `voidvpn disconnect` first.
2. If that fails, delete the state file manually (see paths above).
3. If routes or DNS are still misconfigured, see the DNS and route
   troubleshooting sections.

### "no private key found"

The connect command looks for a private key stored under the server name, then
under "default".  Solutions:

- Generate and save a key: `voidvpn keygen --save --name <server>`
- Import a .conf file that contains the private key:
  `voidvpn servers import <file.conf>`

### TUN device creation fails

- **Windows:** Ensure `wintun.dll` is in the same directory as `voidvpn.exe`.
  Verify you are running as Administrator.
- **Linux:** Ensure `/dev/net/tun` exists and is accessible.  On minimal
  containers you may need to `mkdir -p /dev/net && mknod /dev/net/tun c 10 200`.

### Routes not removed after crash

If the 0.0.0.0/1 and 128.0.0.0/1 routes remain after a crash:

- **Windows:**

      route delete 0.0.0.0 mask 128.0.0.0
      route delete 128.0.0.0 mask 128.0.0.0

- **Linux:**

      sudo ip route delete 0.0.0.0/1
      sudo ip route delete 128.0.0.0/1

---

## Windows-Specific Notes

### Wintun driver

VoidVPN uses the Wintun driver for TUN device creation on Windows.  The
`wintun.dll` file must be placed in the same directory as `voidvpn.exe`.  It is
included in release archives.  Wintun is a kernel-mode driver developed by the
WireGuard project and is signed by WireGuard LLC.

### Elevated terminal

All VPN operations must be run from an elevated (Administrator) terminal.
Non-elevated operations (keygen, config, server management) work from a
standard terminal.

To check if your terminal is elevated, run `voidvpn connect` -- it will
immediately tell you if privileges are insufficient.

### IPC on Windows

Windows uses TCP on `127.0.0.1:41820` for IPC rather than Unix domain sockets.
This is secured with a randomly generated 64-character hex token stored at
`%APPDATA%\VoidVPN\state\ipc.token`.  The token file is readable only by the
user who created it (file mode 0600).

If port 41820 is in use by another application, VoidVPN will fail to start the
IPC server.  The VPN connection itself will still work, but `voidvpn status` and
`voidvpn disconnect` will not function.  In this case, terminate the VPN process
directly (Ctrl+C or Task Manager).

### Configuration location

All configuration files are stored under `%APPDATA%\VoidVPN\`:

    %APPDATA%\VoidVPN\
      config.yaml
      servers\
        myserver.yaml
      state\
        connection.json
        ipc.token
      keys\
        .salt
        default.key

### DNS management

DNS is configured via `netsh` commands.  VoidVPN saves the current DNS
configuration before modifying it and restores DNS to DHCP mode on disconnect.
If DNS was originally set to static addresses (not DHCP), the restoration will
switch to DHCP, which may not preserve your original static DNS settings.

---

## Linux-Specific Notes

### Root requirement

All VPN operations require root privileges.  Use `sudo`:

    sudo voidvpn connect myserver
    sudo voidvpn disconnect
    sudo voidvpn status

Configuration and key management can be done as a regular user, but the config
files will be owned by that user.  When running `connect` as root, ensure the
config directory is readable by root (it typically is, since root can read
all files).

### /etc/resolv.conf management

VoidVPN directly overwrites `/etc/resolv.conf` with tunnel DNS servers and
restores the original contents on disconnect.  This approach has limitations:

- If your system uses `systemd-resolved` or `resolvconf`, direct writes to
  `/etc/resolv.conf` may be overwritten by the system resolver.
- If `/etc/resolv.conf` is a symlink (common with systemd-resolved), VoidVPN
  will replace the symlink with a regular file.

For systems using `systemd-resolved`, consider temporarily disabling it while
VoidVPN is active, or configure your server's DNS through `systemd-resolved`
directly.

### IPC socket location

The IPC socket is created at `$XDG_RUNTIME_DIR/voidvpn.sock`.  If
`$XDG_RUNTIME_DIR` is not set, it falls back to `/tmp/voidvpn-<uid>/voidvpn.sock`.

The socket is created with restrictive permissions (umask 0077) so only the
owning user can connect.

### Configuration location

    ~/.config/voidvpn/
      config.yaml
      servers/
        myserver.yaml
      state/
        connection.json
      keys/
        .salt
        default.key

Respects `$XDG_CONFIG_HOME` if set (defaults to `~/.config`).

### TUN device

Linux uses the standard kernel TUN driver via `/dev/net/tun`.  The interface is
named `voidvpn0`.  After creation, VoidVPN assigns the IP address and brings the
interface up using `ip addr add` and `ip link set up`.

### Firewall considerations

If you use `iptables`, `nftables`, or `ufw`, ensure that:

- UDP traffic to the WireGuard server endpoint port (typically 51820) is allowed
  on the OUTPUT chain.
- Traffic on the `voidvpn0` interface is allowed.

Example for `ufw`:

    sudo ufw allow out to <server-ip> port 51820 proto udp

---

## CLI Command Reference

    voidvpn                              Show help
    voidvpn connect <server>             Connect to a VPN server
    voidvpn connect <server> --daemon    Connect in background mode
    voidvpn disconnect                   Disconnect active VPN
    voidvpn status                       Show connection status
    voidvpn status --watch               Live-updating status
    voidvpn status --json                JSON output
    voidvpn servers list                 List configured servers
    voidvpn servers add <name>           Add a server (with flags)
    voidvpn servers remove <name>        Remove a server
    voidvpn servers import <file>        Import WireGuard .conf file
    voidvpn keygen                       Generate WireGuard keypair
    voidvpn keygen --save                Generate and save to keystore
    voidvpn keygen --save --name <n>     Save with specific name
    voidvpn config show                  Show current configuration
    voidvpn config set <key> <value>     Set a configuration value
    voidvpn version                      Show version information

Global flags:

    -v, --verbose     Enable verbose output

---

## Shell Completions

Generate shell completion scripts:

    make completions

This produces:

- `completions/voidvpn.bash` -- Bash completions
- `completions/_voidvpn` -- Zsh completions
- `completions/voidvpn.fish` -- Fish completions
- `completions/voidvpn.ps1` -- PowerShell completions

Or generate individually:

    voidvpn completion bash > /etc/bash_completion.d/voidvpn
    voidvpn completion zsh > "${fpath[1]}/_voidvpn"
    voidvpn completion fish > ~/.config/fish/completions/voidvpn.fish
    voidvpn completion powershell > voidvpn.ps1
