package main

import (
	"embed"
	"packwiz-web/internal/config"
	"packwiz-web/internal/database"
	"packwiz-web/internal/server"
)

//go:embed public/*
var publicFiles embed.FS

func main() {
	database.InitDb()

	if config.C.Mode == "development" {
		database.SeedDebugData()
	}

	server.Start(&publicFiles)
}
