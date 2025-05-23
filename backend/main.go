package main

import (
	"packwiz-web/commands"
	"packwiz-web/internal/config"
)

// VersionTag
// this must be exported to set it from "go build" command, but should not be
// accessed directly
var VersionTag string

func main() {
	config.SetVersionTag(VersionTag)
	commands.Execute()
}
