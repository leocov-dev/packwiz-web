package public

import "embed"

//go:embed frontend/*
var frontendFiles embed.FS

func GetFrontendFiles() *embed.FS {
	return &frontendFiles
}
