package packwiz_cli_disabled

import (
	"fmt"
	"path/filepath"
	"regexp"
)

var re = regexp.MustCompile(`Project "([^"]+)" successfully added! \(([^)]+)\)`)

func AddCurseforgeMod(
	modpack,
	name,
	addonId,
	category,
	fileId,
	game string,
) (string, error) {

	args := []string{
		"curseforge",
		"add",
		name,
	}

	if addonId != "" {
		args = append(args, "--addon-id", addonId)
	}

	if category != "" {
		args = append(args, "--category", category)
	}

	if fileId != "" {
		args = append(args, "--file-id", fileId)
	}

	if game != "" {
		args = append(args, "--game", game)
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

func ExportCurseforgePack(modpack, outputDir, side string) (string, error) {
	args := []string{
		"curseforge",
		"export",
		"--output", filepath.Join(outputDir, fmt.Sprintf("%s.zip", modpack)),
		"--side", side,
	}

	output, err := runCommand(modpack, args...)
	if err != nil {
		return "", err
	}

	matches := re.FindStringSubmatch(output)

	if len(matches) != 3 {
		return "", fmt.Errorf("failed to parse output: %s", output)
	}

	return matches[2], nil
}
