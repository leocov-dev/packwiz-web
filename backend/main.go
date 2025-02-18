package main

import (
	"embed"
	"packwiz-web/internal/database"
	"packwiz-web/internal/server"
)

//go:embed public/*
var publicFiles embed.FS

func main() {
	database.InitDb()

	server.Start(&publicFiles)
}
