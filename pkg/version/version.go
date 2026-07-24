// Package version provides build-time injected metadata.
//
// Version, Commit, and Date are set at build time via -ldflags.
// When running with "go run" (no ldflags), they fall back to "dev".
package version

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"strings"
)

// Build-time injected via Makefile's -ldflags.
var (
	Version = "dev"     // semantic version, e.g. "0.1.0"
	Branch  = "unknown" // git branch name
	Commit  = "none"    // git commit hash
	Date    = "unknown" // build timestamp (RFC3339)
)

// Info returns a multi-line summary of the build.
func Info() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Version    %s\n", Version))
	sb.WriteString(fmt.Sprintf("Branch     %s\n", Branch))
	sb.WriteString(fmt.Sprintf("Commit     %s\n", Commit))
	sb.WriteString(fmt.Sprintf("Built at   %s\n", Date))
	sb.WriteString(fmt.Sprintf("Go version %s\n", runtime.Version()))
	sb.WriteString(fmt.Sprintf("OS/Arch    %s/%s\n", runtime.GOOS, runtime.GOARCH))

	if bi, ok := debug.ReadBuildInfo(); ok {
		sb.WriteString(fmt.Sprintf("Module     %s@%s\n", bi.Main.Path, bi.Main.Version))
	}

	return sb.String()
}

// Short returns a one-line summary suitable for logs.
func Short() string {
	return fmt.Sprintf("%s (branch=%s commit=%s built=%s)", Version, Branch, Commit, Date)
}
