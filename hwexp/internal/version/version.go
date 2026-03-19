package version

import "runtime"

// Injected at build time via -ldflags.
var (
	Version   = "dev"
	BuildTime = "unknown"
)

// GoVersion returns the Go runtime version used to build the binary.
func GoVersion() string {
	return runtime.Version()
}
