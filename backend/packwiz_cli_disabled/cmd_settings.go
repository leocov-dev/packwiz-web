package packwiz_cli_disabled

import "strings"

// SetAcceptableVersions set a list of minecraft version strings that are acceptable
// for this modpack
func SetAcceptableVersions(modpack string, versions ...string) error {
	_, err := runCommand(
		modpack,
		"settings",
		"acceptable-versions",
		strings.Join(versions, ","),
	)
	return err
}
