package openvpn

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"
)

func TestNewManagementClient(t *testing.T) {
	client := NewManagementClient(12345)
	if client.addr != "127.0.0.1:12345" {
		t.Errorf("addr = %q, want %q", client.addr, "127.0.0.1:12345")
	}
}

func TestGetStatsParse(t *testing.T) {
	// Start a mock TCP server that responds to "status" command
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Listen error: %v", err)
	}
	defer listener.Close()

	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		conn.SetDeadline(time.Now().Add(3 * time.Second))

		// Send greeting
		fmt.Fprintf(conn, ">INFO:OpenVPN Management Interface Version 1\n")

		// Read the "status" command
		buf := make([]byte, 256)
		conn.Read(buf)

		// Send status response
		fmt.Fprintf(conn, "OpenVPN STATISTICS\n")
		fmt.Fprintf(conn, "Updated,2024-01-01 00:00:00\n")
		fmt.Fprintf(conn, "TUN/TAP read bytes,12345\n")
		fmt.Fprintf(conn, "TUN/TAP write bytes,67890\n")
		fmt.Fprintf(conn, "END\n")
	}()

	// Extract port from listener address
	addr := listener.Addr().(*net.TCPAddr)
	client := NewManagementClient(addr.Port)

	stats, err := client.GetStats()
	if err != nil {
		t.Fatalf("GetStats() error: %v", err)
	}
	if stats.RxBytes != 12345 {
		t.Errorf("RxBytes = %d, want 12345", stats.RxBytes)
	}
	if stats.TxBytes != 67890 {
		t.Errorf("TxBytes = %d, want 67890", stats.TxBytes)
	}
}

func TestSendSignal(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Listen error: %v", err)
	}
	defer listener.Close()

	received := make(chan string, 1)
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		conn.SetDeadline(time.Now().Add(3 * time.Second))

		// Send greeting
		fmt.Fprintf(conn, ">INFO:OpenVPN Management Interface Version 1\n")

		// Read the signal command
		buf := make([]byte, 256)
		n, _ := conn.Read(buf)
		received <- string(buf[:n])

		// Send response
		fmt.Fprintf(conn, "SUCCESS: signal SIGTERM thrown\n")
	}()

	addr := listener.Addr().(*net.TCPAddr)
	client := NewManagementClient(addr.Port)

	err = client.SendSignal("SIGTERM")
	if err != nil {
		t.Fatalf("SendSignal() error: %v", err)
	}

	select {
	case cmd := <-received:
		if !strings.Contains(cmd, "signal SIGTERM") {
			t.Errorf("received = %q, want to contain 'signal SIGTERM'", cmd)
		}
	case <-time.After(2 * time.Second):
		t.Error("timeout waiting for signal command")
	}
}

func TestNewManagementClientPort(t *testing.T) {
	tests := []struct {
		port int
		want string
	}{
		{0, "127.0.0.1:0"},
		{65535, "127.0.0.1:65535"},
		{12345, "127.0.0.1:12345"},
	}
	for _, tt := range tests {
		mc := NewManagementClient(tt.port)
		if mc.addr != tt.want {
			t.Errorf("NewManagementClient(%d).addr = %q, want %q", tt.port, mc.addr, tt.want)
		}
	}
}

func TestGetStatsZeroBytes(t *testing.T) {
	// Start a mock management server
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer listener.Close()

	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		conn.SetDeadline(time.Now().Add(3 * time.Second))

		// Send greeting (required by GetStats which reads it first)
		fmt.Fprintf(conn, ">INFO:OpenVPN Management Interface Version 1\n")

		reader := bufio.NewReader(conn)
		line, _ := reader.ReadString('\n')
		if strings.TrimSpace(line) == "status" {
			fmt.Fprintf(conn, "OpenVPN STATISTICS\n")
			fmt.Fprintf(conn, "Updated,2024-01-01 00:00:00\n")
			fmt.Fprintf(conn, "TUN/TAP read bytes,0\n")
			fmt.Fprintf(conn, "TUN/TAP write bytes,0\n")
			fmt.Fprintf(conn, "END\n")
		}
	}()

	addr := listener.Addr().(*net.TCPAddr)
	mc := NewManagementClient(addr.Port)
	stats, err := mc.GetStats()
	if err != nil {
		t.Fatalf("GetStats() error: %v", err)
	}
	if stats.TxBytes != 0 {
		t.Errorf("TxBytes = %d, want 0", stats.TxBytes)
	}
	if stats.RxBytes != 0 {
		t.Errorf("RxBytes = %d, want 0", stats.RxBytes)
	}
}

func TestGetStatsLargeValues(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer listener.Close()

	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		conn.SetDeadline(time.Now().Add(3 * time.Second))

		// Send greeting (required by GetStats which reads it first)
		fmt.Fprintf(conn, ">INFO:OpenVPN Management Interface Version 1\n")

		reader := bufio.NewReader(conn)
		line, _ := reader.ReadString('\n')
		if strings.TrimSpace(line) == "status" {
			fmt.Fprintf(conn, "OpenVPN STATISTICS\n")
			fmt.Fprintf(conn, "Updated,2024-01-01 00:00:00\n")
			fmt.Fprintf(conn, "TUN/TAP read bytes,9999999999\n")
			fmt.Fprintf(conn, "TUN/TAP write bytes,8888888888\n")
			fmt.Fprintf(conn, "END\n")
		}
	}()

	addr := listener.Addr().(*net.TCPAddr)
	mc := NewManagementClient(addr.Port)
	stats, err := mc.GetStats()
	if err != nil {
		t.Fatalf("GetStats() error: %v", err)
	}
	if stats.RxBytes != 9999999999 {
		t.Errorf("RxBytes = %d, want 9999999999", stats.RxBytes)
	}
	if stats.TxBytes != 8888888888 {
		t.Errorf("TxBytes = %d, want 8888888888", stats.TxBytes)
	}
}
