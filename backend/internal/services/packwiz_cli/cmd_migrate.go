package packwiz_cli

// MigrateLoader migrate the loader version in this modpack to the new target
// the target can be a version string or the literals "latest", "recommended"
func MigrateLoader(modpack, target string) error {
	return runCommand(modpack, "migrate", "loader", target)
}

// MigrateMinecraft migrate the minecraft version in this modpack to the new
// version target
func MigrateMinecraft(modpack, version string) error {
	return runCommand(modpack, "migrate", "minecraft", version)
}
