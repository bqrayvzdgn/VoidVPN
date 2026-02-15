//go:build windows

package daemon

import "syscall"

const processQueryLimitedInformation = 0x1000

func isProcessRunning(pid int) bool {
	handle, err := syscall.OpenProcess(processQueryLimitedInformation, false, uint32(pid))
	if err != nil {
		return false
	}
	syscall.CloseHandle(handle)
	return true
}
