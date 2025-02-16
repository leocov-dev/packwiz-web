package packwiz_cli

func RemoveMod(modpack, name string) error {
	return runCommand(modpack, "remove", name)
}
