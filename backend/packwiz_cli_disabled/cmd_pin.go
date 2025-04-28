package packwiz_cli_disabled

// PinMod pin a named mod at its current version
func PinMod(modpack, name string) error {
	_, err := runCommand(modpack, "pin", name)
	return err
}

// UnpinMod unpin a mod so that the tool can update it freely
func UnpinMod(modpack, name string) error {
	_, err := runCommand(modpack, "unpin", name)
	return err
}
