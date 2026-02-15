package openvpn

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

// TrafficStats holds bytes transferred through the tunnel.
type TrafficStats struct {
	TxBytes int64
	RxBytes int64
}

// ManagementClient communicates with the OpenVPN management interface.
type ManagementClient struct {
	addr string
}

// NewManagementClient creates a client for the given management address.
func NewManagementClient(port int) *ManagementClient {
	return &ManagementClient{
		addr: fmt.Sprintf("127.0.0.1:%d", port),
	}
}

// GetStats queries the management interface for traffic statistics.
func (m *ManagementClient) GetStats() (*TrafficStats, error) {
	conn, err := net.DialTimeout("tcp", m.addr, 2*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to management interface: %w", err)
	}
	defer conn.Close()

	conn.SetDeadline(time.Now().Add(3 * time.Second))

	reader := bufio.NewReader(conn)

	// Read the greeting line
	reader.ReadString('\n')

	// Send status command
	fmt.Fprintf(conn, "status\n")

	stats := &TrafficStats{}
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		line = strings.TrimSpace(line)
		if line == "END" {
			break
		}

		// Parse "TUN/TAP read bytes,<N>" and "TUN/TAP write bytes,<N>"
		if strings.HasPrefix(line, "TUN/TAP read bytes,") {
			parts := strings.SplitN(line, ",", 2)
			if len(parts) == 2 {
				if n, err := strconv.ParseInt(parts[1], 10, 64); err == nil {
					stats.RxBytes = n
				}
			}
		}
		if strings.HasPrefix(line, "TUN/TAP write bytes,") {
			parts := strings.SplitN(line, ",", 2)
			if len(parts) == 2 {
				if n, err := strconv.ParseInt(parts[1], 10, 64); err == nil {
					stats.TxBytes = n
				}
			}
		}
	}

	return stats, nil
}

// SendSignal sends a signal command (e.g., SIGTERM) via the management interface.
func (m *ManagementClient) SendSignal(sig string) error {
	conn, err := net.DialTimeout("tcp", m.addr, 2*time.Second)
	if err != nil {
		return fmt.Errorf("failed to connect to management interface: %w", err)
	}
	defer conn.Close()

	conn.SetDeadline(time.Now().Add(3 * time.Second))

	reader := bufio.NewReader(conn)
	reader.ReadString('\n') // greeting

	fmt.Fprintf(conn, "signal %s\n", sig)

	// Read response
	reader.ReadString('\n')
	return nil
}
