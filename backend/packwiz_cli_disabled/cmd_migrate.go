package packwiz_cli_disabled

// MigrateLoader migrate the loader version in this modpack to the new target
// the target can be a version string or the literals "latest", "recommended"
func MigrateLoader(modpack, target string) error {
	_, err := runCommand(modpack, "migrate", "loader", target)
	return err
}

// MigrateMinecraft migrate the minecraft version in this modpack to the new
// version target
func MigrateMinecraft(modpack, version string) error {
	_, err := runCommand(modpack, "migrate", "minecraft", version)
	return err
}
