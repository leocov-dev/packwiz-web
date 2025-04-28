package packwiz_cli_disabled

func RemoveMod(modpack, name string) error {
	_, err := runCommand(modpack, "remove", name)
	return err
}
