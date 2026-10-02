package vek

import "runtime/debug"

// Version can be set with -ldflags "-X github.com/vekio/vek/internal/cli/vek.Version=vX.Y.Z".
var Version string

func buildVersion() string {
	// Prefer a version supplied when building release binaries.
	if Version != "" {
		return Version
	}
	// Go embeds the module version when installing from a tag or commit.
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	return "dev"
}
