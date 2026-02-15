package daemon

import (
	"testing"
)

func TestMarshalUnmarshalRequest(t *testing.T) {
	data, err := MarshalRequest("status")
	if err != nil {
		t.Fatalf("MarshalRequest() error: %v", err)
	}

	req, err := UnmarshalRequest(data)
	if err != nil {
		t.Fatalf("UnmarshalRequest() error: %v", err)
	}

	if req.Command != "status" {
		t.Errorf("Command = %q, want %q", req.Command, "status")
	}
}

func TestMarshalUnmarshalResponse(t *testing.T) {
	resp := &IPCResponse{
		Success: true,
		State: &ConnectionState{
			Server:   "test-server",
			TxBytes:  100,
			RxBytes:  200,
		},
	}

	data, err := MarshalResponse(resp)
	if err != nil {
		t.Fatalf("MarshalResponse() error: %v", err)
	}

	loaded, err := UnmarshalResponse(data)
	if err != nil {
		t.Fatalf("UnmarshalResponse() error: %v", err)
	}

	if !loaded.Success {
		t.Error("Success should be true")
	}
	if loaded.State.Server != "test-server" {
		t.Errorf("State.Server = %q, want %q", loaded.State.Server, "test-server")
	}
}

func TestUnmarshalRequestInvalid(t *testing.T) {
	_, err := UnmarshalRequest([]byte("not json"))
	if err == nil {
		t.Error("should error on invalid JSON")
	}
}

func TestUnmarshalResponseInvalid(t *testing.T) {
	_, err := UnmarshalResponse([]byte("not json"))
	if err == nil {
		t.Error("should error on invalid JSON")
	}
}

func TestResponseWithError(t *testing.T) {
	resp := &IPCResponse{
		Success: false,
		Error:   "something went wrong",
	}

	data, _ := MarshalResponse(resp)
	loaded, _ := UnmarshalResponse(data)

	if loaded.Success {
		t.Error("Success should be false")
	}
	if loaded.Error != "something went wrong" {
		t.Errorf("Error = %q, want %q", loaded.Error, "something went wrong")
	}
}
