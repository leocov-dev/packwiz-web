package packwiz_cli

func AddCurseforgeMod(
	modpack,
	name,
	addonId,
	category,
	fileId,
	game string) error {

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

	return runCommand(modpack, args...)
}
