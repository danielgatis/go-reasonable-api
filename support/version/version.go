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
	return fmt.Sprintf("%s (commit: %s, built: %s, go: %s)",
		Version, Commit, BuildDate, runtime.Version())
}

func Short() string {
	return Version
}
