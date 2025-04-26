package packwiz_cli

// UpdateOne update a named mod in the active modpack
func UpdateOne(modpack, name string) error {
	_, err := runCommand(modpack, "update", name)
	return err
}

// UpdateAll update all mods in the active modpack
func UpdateAll(modpack string) error {
	_, err := runCommand(modpack, "update", "--all")
	return err
}
