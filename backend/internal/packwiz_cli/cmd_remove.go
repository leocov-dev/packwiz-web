package packwiz_cli

func RemoveMod(modpack, name string) error {
	_, err := runCommand(modpack, "remove", name)
	return err
}
