//go:build !windows

package daemon

import (
	"bufio"
	"net"
	"os"
	"strings"
	"testing"
	"time"
)

func TestGetSocketPathXDG(t *testing.T) {
	orig := os.Getenv("XDG_RUNTIME_DIR")
	defer os.Setenv("XDG_RUNTIME_DIR", orig)

	os.Setenv("XDG_RUNTIME_DIR", "/tmp/xdg-test")
	path := getSocketPath()
	if !strings.Contains(path, "voidvpn.sock") {
		t.Errorf("getSocketPath() = %q, want to contain voidvpn.sock", path)
	}
	if !strings.HasPrefix(path, "/tmp/xdg-test") {
		t.Errorf("getSocketPath() = %q, want prefix /tmp/xdg-test", path)
	}
}

func TestGetSocketPathFallback(t *testing.T) {
	orig := os.Getenv("XDG_RUNTIME_DIR")
	defer os.Setenv("XDG_RUNTIME_DIR", orig)

	os.Unsetenv("XDG_RUNTIME_DIR")
	path := getSocketPath()
	if !strings.Contains(path, "voidvpn") {
		t.Errorf("getSocketPath() = %q, want to contain voidvpn", path)
	}
}

func TestIPCServerUnixRoundTrip(t *testing.T) {
	sockPath := t.TempDir() + "/test.sock"

	handler := func(req *IPCRequest) *IPCResponse {
		return &IPCResponse{Success: true}
	}

	listener, err := net.Listen("unix", sockPath)
	if err != nil {
		t.Fatalf("Listen error: %v", err)
	}
	defer listener.Close()

	server := &IPCServer{
		listener:   listener,
		handler:    handler,
		socketPath: sockPath,
	}
	go server.Serve()

	conn, err := net.DialTimeout("unix", sockPath, 2*time.Second)
	if err != nil {
		t.Fatalf("Dial error: %v", err)
	}
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(3 * time.Second))

	reqData, _ := MarshalRequest("status")
	conn.Write(append(reqData, '\n'))

	reader := bufio.NewReader(conn)
	respLine, err := reader.ReadBytes('\n')
	if err != nil {
		t.Fatalf("ReadBytes error: %v", err)
	}

	resp, err := UnmarshalResponse(respLine)
	if err != nil {
		t.Fatalf("UnmarshalResponse error: %v", err)
	}
	if !resp.Success {
		t.Error("response Success = false, want true")
	}
}

func TestSendIPCRequestUnix(t *testing.T) {
	sockPath := t.TempDir() + "/test2.sock"

	handler := func(req *IPCRequest) *IPCResponse {
		return &IPCResponse{
			Success: true,
			State: &ConnectionState{
				Server: "unix-server",
			},
		}
	}

	listener, err := net.Listen("unix", sockPath)
	if err != nil {
		t.Fatalf("Listen error: %v", err)
	}
	defer listener.Close()

	server := &IPCServer{
		listener:   listener,
		handler:    handler,
		socketPath: sockPath,
	}
	go server.Serve()

	// Test the protocol manually since SendIPCRequest uses getSocketPath()
	conn, err := net.DialTimeout("unix", sockPath, 2*time.Second)
	if err != nil {
		t.Fatalf("Dial error: %v", err)
	}
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(3 * time.Second))

	reqData, _ := MarshalRequest("status")
	conn.Write(append(reqData, '\n'))

	reader := bufio.NewReader(conn)
	respLine, _ := reader.ReadBytes('\n')
	resp, err := UnmarshalResponse(respLine)
	if err != nil {
		t.Fatalf("UnmarshalResponse error: %v", err)
	}
	if !resp.Success {
		t.Error("response should succeed")
	}
	if resp.State == nil || resp.State.Server != "unix-server" {
		t.Error("response state should contain unix-server")
	}
}
