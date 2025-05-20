package packwiz_svc

import (
	"fmt"
)

func getModSourceLink(source, key, version string) string {
	switch source {
	case "modrinth":
		return fmt.Sprintf("https://modrinth.com/mod/%s/version/%s", key, version)
	case "curseforge":
		return fmt.Sprintf("https://www.curseforge.com/minecraft/mc-mods/%s/files/%s", key, version)
	default:
		return ""
	}
}
