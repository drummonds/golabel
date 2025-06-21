package version

import (
	"fmt"
	"time"
)

// Version information
var (
	Version   = "0.1.0"
	BuildDate = time.Now().Format("2006-01-02 15:04:05")
	GitCommit = "golabel"
)

// GetVersionInfo returns formatted version information
func GetVersionInfo() string {
	return fmt.Sprintf("v%s (%s) - %s", Version, GitCommit, BuildDate)
}

// GetVersion returns just the version string
func GetVersion() string {
	return Version
}

// GetBuildDate returns the build date
func GetBuildDate() string {
	return BuildDate
}
