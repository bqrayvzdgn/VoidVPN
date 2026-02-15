package openvpn

import (
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
