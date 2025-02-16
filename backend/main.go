package main

import (
	"packwiz-web/internal/database"
	"packwiz-web/internal/server"
)

func main() {
	database.InitDb()

	server.Start()
}
