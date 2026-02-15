//go:build windows

package daemon

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type IPCServer struct {
	listener net.Listener
	handler  func(*IPCRequest) *IPCResponse
	token    string
}

func ipcTokenPath() string {
	return filepath.Join(os.Getenv("APPDATA"), "VoidVPN", "state", "ipc.token")
}

func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func NewIPCServer(handler func(*IPCRequest) *IPCResponse) (*IPCServer, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:41820")
	if err != nil {
		return nil, fmt.Errorf("failed to start IPC server: %w", err)
	}

	// Generate and save auth token
	token, err := generateToken()
	if err != nil {
		listener.Close()
		return nil, fmt.Errorf("failed to generate IPC token: %w", err)
	}

	tokenDir := filepath.Dir(ipcTokenPath())
	if err := os.MkdirAll(tokenDir, 0700); err != nil {
		listener.Close()
		return nil, err
	}
	if err := os.WriteFile(ipcTokenPath(), []byte(token), 0600); err != nil {
		listener.Close()
		return nil, err
	}

	return &IPCServer{
		listener: listener,
		handler:  handler,
		token:    token,
	}, nil
}

func (s *IPCServer) Serve() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			return
		}
		go s.handleConn(conn)
	}
}

func (s *IPCServer) handleConn(conn net.Conn) {
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(5 * time.Second))

	reader := bufio.NewReader(conn)

	// Read auth token line first
	tokenLine, err := reader.ReadString('\n')
	if err != nil {
		return
	}
	if strings.TrimSpace(tokenLine) != s.token {
		resp := &IPCResponse{Success: false, Error: "authentication failed"}
		respData, _ := MarshalResponse(resp)
		conn.Write(append(respData, '\n'))
		return
	}

	// Read command
	data, err := reader.ReadBytes('\n')
	if err != nil && err != io.EOF {
		return
	}

	req, err := UnmarshalRequest(data)
	if err != nil {
		resp := &IPCResponse{Success: false, Error: err.Error()}
		respData, _ := MarshalResponse(resp)
		conn.Write(append(respData, '\n'))
		return
	}

	resp := s.handler(req)
	respData, _ := MarshalResponse(resp)
	conn.Write(append(respData, '\n'))
}

func (s *IPCServer) Close() error {
	os.Remove(ipcTokenPath())
	return s.listener.Close()
}

func SendIPCRequest(cmd string) (*IPCResponse, error) {
	// Read auth token
	tokenData, err := os.ReadFile(ipcTokenPath())
	if err != nil {
		return nil, fmt.Errorf("VPN is not running (no auth token found): %w", err)
	}
	token := strings.TrimSpace(string(tokenData))

	conn, err := net.DialTimeout("tcp", "127.0.0.1:41820", 3*time.Second)
	if err != nil {
		return nil, fmt.Errorf("VPN is not running (could not connect to IPC): %w", err)
	}
	defer conn.Close()

	conn.SetDeadline(time.Now().Add(5 * time.Second))

	// Send auth token
	if _, err := fmt.Fprintf(conn, "%s\n", token); err != nil {
		return nil, err
	}

	// Send command
	reqData, err := MarshalRequest(cmd)
	if err != nil {
		return nil, err
	}
	if _, err := conn.Write(append(reqData, '\n')); err != nil {
		return nil, err
	}

	reader := bufio.NewReader(conn)
	respData, err := reader.ReadBytes('\n')
	if err != nil && err != io.EOF {
		return nil, err
	}

	return UnmarshalResponse(respData)
}
