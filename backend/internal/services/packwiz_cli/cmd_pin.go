package packwiz_cli

// PinMod pin a named mod at its current version
func PinMod(modpack, name string) error {
	return runCommand(modpack, "pin", name)
}

// UnpinMod unpin a mod so that the tool can update it freely
func UnpinMod(modpack, name string) error {
	return runCommand(modpack, "unpin", name)
}
