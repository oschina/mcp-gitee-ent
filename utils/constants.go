package utils

import (
	"runtime/debug"
	"strings"
)

// Version is injected by release builds. Go module installs fall back to the
// module version recorded in the binary's build information.
var Version = "dev"

func init() {
	if Version != "dev" {
		return
	}
	info, ok := debug.ReadBuildInfo()
	if !ok || info.Main.Version == "" || info.Main.Version == "(devel)" {
		return
	}
	Version = strings.TrimPrefix(info.Main.Version, "v")
}
