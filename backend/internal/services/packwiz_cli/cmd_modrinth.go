package packwiz_cli

import (
	"fmt"
	"path/filepath"
)

func AddModrinthMod(
	modpack,
	name,
	projectId,
	versionFilename,
	versionId string,
) error {
	args := []string{
		"modrinth",
		"add",
		name,
	}

	if projectId != "" {
		args = append(args, "--project-id", projectId)
	}
	if versionFilename != "" {
		args = append(args, "--version-filename", versionFilename)
	}
	if versionId != "" {
		args = append(args, "--version-id", versionId)
	}

	return runCommand(modpack, args...)
}

func ExportModrinthPack(modpack, outputDir string, restrictDomains bool) error {
	args := []string{
		"modrinth",
		"export",
		"--output", filepath.Join(outputDir, fmt.Sprintf("%s.zip", modpack)),
	}

	if restrictDomains {
		args = append(args, "--restrict-domains", "true")
	} else {
		args = append(args, "--restrict-domains", "false")
	}

	return runCommand(modpack, args...)
}
