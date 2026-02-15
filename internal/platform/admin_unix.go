//go:build !windows

package platform

import "os"

func isAdmin() bool {
	return os.Geteuid() == 0
}
