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
) (string, error) {
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

	output, err := runCommand(modpack, args...)
	if err != nil {
		return "", err
	}

	matches := re.FindStringSubmatch(output)

	if len(matches) != 3 {
		return "", fmt.Errorf("failed to parse output: %s", output)
	}

	return matches[1], nil
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

	_, err := runCommand(modpack, args...)
	return err
}
