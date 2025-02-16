package packwiz_cli

// UpdateOne update a named mod in the active modpack
func UpdateOne(modpack, name string) error {
	return runCommand(modpack, "update", name)
}

// UpdateAll update all mods in the active modpack
func UpdateAll(modpack string) error {
	return runCommand(modpack, "update", "--all")
}
