package buildinfo

import "runtime"

var (
	Version   = "dev"
	Commit    = "unknown"
	BuildDate = "unknown"
)

func GoVersion() string {
	return runtime.Version()
}

func Short() string {
	return "gate-limiter " + Version
}

func Full() string {
	return "gate-limiter " + Version + " (commit: " + Commit + ", built: " + BuildDate + ", " + GoVersion() + ")"
}
