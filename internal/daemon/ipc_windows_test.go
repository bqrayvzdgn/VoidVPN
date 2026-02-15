//go:build windows

package daemon

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"strings"
	"testing"
	"time"
)

func setupIPCTestEnv(t *testing.T) func() {
	t.Helper()
	tmpDir := t.TempDir()
	orig := os.Getenv("APPDATA")
	os.Setenv("APPDATA", tmpDir)
	os.MkdirAll(tmpDir+`\VoidVPN\state`, 0700)
	return func() {
		os.Setenv("APPDATA", orig)
	}
}

func TestGenerateToken(t *testing.T) {
	token, err := generateToken()
	if err != nil {
		t.Fatalf("generateToken() error: %v", err)
	}
	if len(token) != 64 {
		t.Errorf("token length = %d, want 64", len(token))
	}
	// Verify it's valid hex
	if _, err := hex.DecodeString(token); err != nil {
		t.Errorf("token is not valid hex: %v", err)
	}
}

func TestGenerateTokenUnique(t *testing.T) {
	t1, _ := generateToken()
	t2, _ := generateToken()
	if t1 == t2 {
		t.Error("generateToken() returned same value twice")
	}
}

func TestIPCServerAuthSuccess(t *testing.T) {
	cleanup := setupIPCTestEnv(t)
	defer cleanup()

	handler := func(req *IPCRequest) *IPCResponse {
		return &IPCResponse{Success: true}
	}

	// Listen on a random port to avoid conflicts
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Listen error: %v", err)
	}
	defer listener.Close()

	token := "testtoken123"
	server := &IPCServer{
		listener: listener,
		handler:  handler,
		token:    token,
	}
	go server.Serve()

	// Connect and authenticate
	conn, err := net.DialTimeout("tcp", listener.Addr().String(), 2*time.Second)
	if err != nil {
		t.Fatalf("Dial error: %v", err)
	}
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(3 * time.Second))

	// Send token
	fmt.Fprintf(conn, "%s\n", token)
	// Send command
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
		t.Errorf("response Success = false, want true")
	}
}

func TestIPCServerAuthFailure(t *testing.T) {
	cleanup := setupIPCTestEnv(t)
	defer cleanup()

	handler := func(req *IPCRequest) *IPCResponse {
		return &IPCResponse{Success: true}
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Listen error: %v", err)
	}
	defer listener.Close()

	server := &IPCServer{
		listener: listener,
		handler:  handler,
		token:    "correct-token",
	}
	go server.Serve()

	conn, err := net.DialTimeout("tcp", listener.Addr().String(), 2*time.Second)
	if err != nil {
		t.Fatalf("Dial error: %v", err)
	}
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(3 * time.Second))

	// Send wrong token
	fmt.Fprintf(conn, "wrong-token\n")

	reader := bufio.NewReader(conn)
	respLine, err := reader.ReadBytes('\n')
	if err != nil {
		t.Fatalf("ReadBytes error: %v", err)
	}

	resp, err := UnmarshalResponse(respLine)
	if err != nil {
		t.Fatalf("UnmarshalResponse error: %v", err)
	}
	if resp.Success {
		t.Error("response Success = true, want false for wrong token")
	}
	if !strings.Contains(resp.Error, "authentication") {
		t.Errorf("error = %q, want mention of authentication", resp.Error)
	}
}

func TestSendIPCRequestRoundTrip(t *testing.T) {
	cleanup := setupIPCTestEnv(t)
	defer cleanup()

	handler := func(req *IPCRequest) *IPCResponse {
		return &IPCResponse{
			Success: true,
			State: &ConnectionState{
				Server: "test-server",
			},
		}
	}

	// Create a server on a random port and write token
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Listen error: %v", err)
	}
	defer listener.Close()

	token := "test-round-trip-token"
	server := &IPCServer{
		listener: listener,
		handler:  handler,
		token:    token,
	}
	go server.Serve()

	// Write token file
	os.WriteFile(ipcTokenPath(), []byte(token), 0600)

	// We can't use SendIPCRequest directly since it hardcodes port 41820.
	// Instead, test the protocol manually at the dynamic port.
	conn, err := net.DialTimeout("tcp", listener.Addr().String(), 2*time.Second)
	if err != nil {
		t.Fatalf("Dial error: %v", err)
	}
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(3 * time.Second))

	fmt.Fprintf(conn, "%s\n", token)
	reqData, _ := MarshalRequest("status")
	conn.Write(append(reqData, '\n'))

	reader := bufio.NewReader(conn)
	respLine, _ := reader.ReadBytes('\n')
	resp, err := UnmarshalResponse(respLine)
	if err != nil {
		t.Fatalf("UnmarshalResponse error: %v", err)
	}
	if !resp.Success {
		t.Error("round-trip should succeed")
	}
	if resp.State == nil || resp.State.Server != "test-server" {
		t.Error("response state should contain test-server")
	}
}
