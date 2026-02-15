// Package version provides build version information set via ldflags.
package version

import (
	"fmt"
	"runtime"
)

var (
	Version   = "dev"
	Commit    = "none"
	BuildDate = "unknown"
)

func Full() string {
	return fmt.Sprintf("VoidVPN %s (%s) built %s [%s/%s]", Version, Commit, BuildDate, runtime.GOOS, runtime.GOARCH)
}

func Short() string {
	return Version
}
