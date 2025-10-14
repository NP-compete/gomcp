package version

import (
	"fmt"
	"runtime"
	"time"
)

var (
	// Version is the application version
	// Set via ldflags: -X github.com/NP-compete/gomcp/internal/version.Version=x.y.z
	Version = "dev"

	// GitCommit is the git commit hash
	// Set via ldflags: -X github.com/NP-compete/gomcp/internal/version.GitCommit=$(git rev-parse HEAD)
	GitCommit = "unknown"

	// BuildTime is when the binary was built
	// Set via ldflags: -X github.com/NP-compete/gomcp/internal/version.BuildTime=$(date -u '+%Y-%m-%d_%H:%M:%S')
	BuildTime = "unknown"

	// GoVersion is the Go version used to build
	GoVersion = runtime.Version()
)

// Info represents build information
type Info struct {
	Version   string `json:"version"`
	GitCommit string `json:"git_commit"`
	BuildTime string `json:"build_time"`
	GoVersion string `json:"go_version"`
	OS        string `json:"os"`
	Arch      string `json:"arch"`
}

// Get returns version information
func Get() Info {
	return Info{
		Version:   Version,
		GitCommit: GitCommit,
		BuildTime: BuildTime,
		GoVersion: GoVersion,
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
	}
}

// String returns a formatted version string
func String() string {
	info := Get()
	return fmt.Sprintf(
		"gomcp %s (commit: %s, built: %s, go: %s, os: %s, arch: %s)",
		info.Version,
		info.GitCommit[:min(7, len(info.GitCommit))],
		info.BuildTime,
		info.GoVersion,
		info.OS,
		info.Arch,
	)
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// BuildDate returns the parsed build time
func BuildDate() (time.Time, error) {
	return time.Parse("2006-01-02_15:04:05", BuildTime)
}
