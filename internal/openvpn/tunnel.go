package openvpn

import (
	"bufio"
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/voidvpn/voidvpn/internal/config"
	"github.com/voidvpn/voidvpn/internal/tunnel"
)

type scanResult struct {
	connected bool
	errMsg    string
	lastLines []string
}

// Tunnel manages an OpenVPN connection by shelling out to the openvpn binary.
type Tunnel struct {
	server     *config.ServerConfig
	cmd        *exec.Cmd
	mgmt       *ManagementClient
	mgmtPort   int
	tmpConfig  string
	cancel     context.CancelFunc
	mu         sync.Mutex
	connectedAt time.Time
}

// NewTunnel creates a new OpenVPN tunnel for the given server config.
func NewTunnel(serverCfg *config.ServerConfig) *Tunnel {
	port := 10000 + rand.Intn(50000)
	return &Tunnel{
		server:   serverCfg,
		mgmtPort: port,
		mgmt:     NewManagementClient(port),
	}
}

func (t *Tunnel) Connect(ctx context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Detect openvpn binary
	binPath, err := DetectOpenVPN()
	if err != nil {
		return err
	}

	// Write temp config
	tmpDir := os.TempDir()
	t.tmpConfig = filepath.Join(tmpDir, fmt.Sprintf("voidvpn-ovpn-%d.conf", os.Getpid()))
	ovpnContent := BuildOVPNConfig(t.server, t.mgmtPort)
	if err := os.WriteFile(t.tmpConfig, []byte(ovpnContent), 0600); err != nil {
		return fmt.Errorf("failed to write temp config: %w", err)
	}

	ctx, cancel := context.WithCancel(ctx)
	t.cancel = cancel

	t.cmd = exec.CommandContext(ctx, binPath, "--config", t.tmpConfig)

	// Combine stdout and stderr into a single pipe so we capture ALL openvpn output.
	// On Windows, critical messages (TAP adapter errors, etc.) may go to stderr.
	pr, pw, err := os.Pipe()
	if err != nil {
		cancel()
		t.cleanupTempFile()
		return fmt.Errorf("failed to create pipe: %w", err)
	}
	t.cmd.Stdout = pw
	t.cmd.Stderr = pw

	if err := t.cmd.Start(); err != nil {
		cancel()
		pw.Close()
		pr.Close()
		t.cleanupTempFile()
		return fmt.Errorf("failed to start openvpn: %w", err)
	}
	pw.Close() // Close write end in parent; child process has its own fd

	// Debug log file for full output
	logPath := filepath.Join(os.TempDir(), "voidvpn-ovpn-debug.log")
	logFile, _ := os.Create(logPath)

	// Scan combined output in background goroutine.
	// Signals one of: connected, error message, or process exit (EOF).
	resultCh := make(chan scanResult, 1)
	go func() {
		defer pr.Close()
		defer func() {
			if logFile != nil {
				logFile.Close()
			}
		}()
		var lines []string
		scanner := bufio.NewScanner(pr)
		for scanner.Scan() {
			line := scanner.Text()
			if logFile != nil {
				fmt.Fprintln(logFile, line)
			}
			lines = append(lines, line)
			if len(lines) > 10 {
				lines = lines[1:]
			}
			if strings.Contains(line, "Initialization Sequence Completed") {
				resultCh <- scanResult{connected: true}
				return
			}
			if strings.Contains(line, "AUTH_FAILED") || strings.Contains(line, "Connection refused") {
				resultCh <- scanResult{errMsg: line}
				return
			}
		}
		// Scanner finished (EOF) = process exited without connecting
		resultCh <- scanResult{lastLines: lines}
	}()

	select {
	case res := <-resultCh:
		if res.connected {
			t.connectedAt = time.Now()
			return nil
		}
		_ = t.cmd.Process.Kill()
		_ = t.cmd.Wait()
		t.cleanupTempFile()
		cancel()
		if res.errMsg != "" {
			return fmt.Errorf("openvpn error: %s", res.errMsg)
		}
		detail := strings.Join(res.lastLines, "\n  ")
		if detail == "" {
			detail = "no output captured"
		}
		return fmt.Errorf("openvpn exited unexpectedly. Last output:\n  %s\nFull log: %s", detail, logPath)
	case <-time.After(60 * time.Second):
		_ = t.cmd.Process.Kill()
		_ = t.cmd.Wait()
		t.cleanupTempFile()
		cancel()
		return fmt.Errorf("openvpn connection timed out after 60s (check %s)", logPath)
	case <-ctx.Done():
		_ = t.cmd.Process.Kill()
		_ = t.cmd.Wait()
		t.cleanupTempFile()
		return ctx.Err()
	}
}

func (t *Tunnel) Disconnect() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Try graceful shutdown via management interface
	if t.mgmt != nil {
		_ = t.mgmt.SendSignal("SIGTERM")
	}

	// If process is still running, kill it
	if t.cmd != nil && t.cmd.Process != nil {
		_ = t.cmd.Process.Kill()
		_ = t.cmd.Wait()
	}

	if t.cancel != nil {
		t.cancel()
	}

	t.cleanupTempFile()
	return nil
}

func (t *Tunnel) Status() (*tunnel.TunnelStatus, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	status := &tunnel.TunnelStatus{
		Protocol:   "openvpn",
		ServerName: t.server.Name,
		Endpoint:   t.server.Endpoint,
		TunnelIP:   t.server.Address,
		Connected:  t.IsActiveUnlocked(),
	}

	if status.Connected {
		status.ConnectedAt = t.connectedAt
		if stats, err := t.mgmt.GetStats(); err == nil {
			status.TxBytes = stats.TxBytes
			status.RxBytes = stats.RxBytes
		}
		status.InterfaceName = "tun0"
	}

	return status, nil
}

func (t *Tunnel) IsActive() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.IsActiveUnlocked()
}

func (t *Tunnel) IsActiveUnlocked() bool {
	if t.cmd == nil || t.cmd.Process == nil {
		return false
	}
	// Check if process is still running
	if t.cmd.ProcessState != nil {
		return false
	}
	return true
}

func (t *Tunnel) cleanupTempFile() {
	if t.tmpConfig != "" {
		os.Remove(t.tmpConfig)
		t.tmpConfig = ""
	}
}
