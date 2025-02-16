package packwiz_cli

import "strings"

// SetAcceptableVersions set a list of minecraft version strings that are acceptable
// for this modpack
func SetAcceptableVersions(modpack string, versions ...string) error {
	return runCommand(
		modpack,
		"settings",
		"acceptable-versions",
		strings.Join(versions, ","),
	)
}
