package packwiz_cli

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
