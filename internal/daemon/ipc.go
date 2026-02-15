package daemon

import (
	"encoding/json"
	"fmt"
)

type IPCRequest struct {
	Command string `json:"command"` // "status", "disconnect"
}

type IPCResponse struct {
	Success bool             `json:"success"`
	Error   string           `json:"error,omitempty"`
	State   *ConnectionState `json:"state,omitempty"`
}

func MarshalRequest(cmd string) ([]byte, error) {
	req := IPCRequest{Command: cmd}
	return json.Marshal(req)
}

func UnmarshalRequest(data []byte) (*IPCRequest, error) {
	var req IPCRequest
	if err := json.Unmarshal(data, &req); err != nil {
		return nil, fmt.Errorf("invalid IPC request: %w", err)
	}
	return &req, nil
}

func MarshalResponse(resp *IPCResponse) ([]byte, error) {
	return json.Marshal(resp)
}

func UnmarshalResponse(data []byte) (*IPCResponse, error) {
	var resp IPCResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("invalid IPC response: %w", err)
	}
	return &resp, nil
}
