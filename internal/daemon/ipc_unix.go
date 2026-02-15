//go:build !windows

package daemon

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

func getSocketPath() string {
	// Use XDG_RUNTIME_DIR if available (user-specific, proper permissions)
	if xdg := os.Getenv("XDG_RUNTIME_DIR"); xdg != "" {
		return filepath.Join(xdg, "voidvpn.sock")
	}
	// Fallback to a user-specific temp directory
	dir := filepath.Join(os.TempDir(), fmt.Sprintf("voidvpn-%d", os.Getuid()))
	os.MkdirAll(dir, 0700)
	return filepath.Join(dir, "voidvpn.sock")
}

type IPCServer struct {
	listener   net.Listener
	handler    func(*IPCRequest) *IPCResponse
	socketPath string
}

func NewIPCServer(handler func(*IPCRequest) *IPCResponse) (*IPCServer, error) {
	sockPath := getSocketPath()

	// Remove stale socket
	os.Remove(sockPath)

	// Set umask before creating socket to avoid TOCTOU race
	oldMask := syscall.Umask(0077)
	listener, err := net.Listen("unix", sockPath)
	syscall.Umask(oldMask)
	if err != nil {
		return nil, fmt.Errorf("failed to start IPC server: %w", err)
	}

	return &IPCServer{
		listener:   listener,
		handler:    handler,
		socketPath: sockPath,
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
	os.Remove(s.socketPath)
	return s.listener.Close()
}

func SendIPCRequest(cmd string) (*IPCResponse, error) {
	sockPath := getSocketPath()
	conn, err := net.DialTimeout("unix", sockPath, 3*time.Second)
	if err != nil {
		return nil, fmt.Errorf("VPN is not running (could not connect to IPC): %w", err)
	}
	defer conn.Close()

	reqData, err := MarshalRequest(cmd)
	if err != nil {
		return nil, err
	}

	conn.SetDeadline(time.Now().Add(5 * time.Second))
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
