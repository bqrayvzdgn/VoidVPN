// Package platform provides platform-specific utilities for VoidVPN.
package platform

// IsAdmin returns true if the current process has elevated/root privileges.
func IsAdmin() bool {
	return isAdmin()
}
