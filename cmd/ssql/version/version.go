package version

import (
	_ "embed"
	"strings"
)

// Version is the current version of ssql.
// This is embedded from version.txt which is generated from git describe.
var Version = strings.TrimSpace(gitVersion)

//go:embed version.txt
var gitVersion string
