package packwiz_cli

import (
	"packwiz-web/internal/types"
)

type PackFile struct {
	Name       string            `toml:"name"`
	Version    string            `toml:"version"`
	PackFormat string            `toml:"pack-format"`
	Index      PackFileIndexMeta `toml:"index"`
	Versions   PackFileVersions  `toml:"versions"`
	Options    PackFileOptions   `toml:"options"`
}

type IndexFile struct {
	HashFormat string      `toml:"hash-format"`
	Files      []IndexMeta `toml:"files"`
}

type ModFile struct {
	Name     string          `toml:"name"`
	Filename string          `toml:"filename"`
	Side     types.ModSide   `toml:"side"`
	Pin      bool            `toml:"pin"`
	Download DownloadMeta    `toml:"download"`
	Update   UpdateSourceMap `toml:"update"`
}

type PackFileIndexMeta struct {
	File       string `toml:"file"`
	HashFormat string `toml:"hash-format"`
	Hash       string `toml:"hash"`
}

type PackFileVersions struct {
	Minecraft  string `toml:"minecraft"`
	Forge      string `toml:"forge,omitempty"`
	Fabric     string `toml:"fabric,omitempty"`
	LiteLoader string `toml:"liteloader,omitempty"`
	Quilt      string `toml:"quilt,omitempty"`
	NeoForge   string `toml:"neoforge,omitempty"`
}

type PackFileOptions struct {
	AcceptableGameVersions []string `toml:"acceptable-game-versions,omitempty"`
}

type IndexMeta struct {
	File     string `toml:"file"`
	Hash     string `toml:"hash"`
	Metafile bool   `toml:"metafile"`
}

type ModrinthMeta struct {
	ModId   string `toml:"mod-id"`
	Version string `toml:"version"`
}

type CurseforgeMeta struct {
	FileId    int `toml:"file-id"`
	ProjectId int `toml:"project-id"`
}

type UpdateSourceMap struct {
	Modrinth   ModrinthMeta   `json:"modrinth"`
	Curseforge CurseforgeMeta `json:"curseforge"`
}

type DownloadMeta struct {
	Url        string `json:"url"`
	Mode       string `json:"mode"`
	Hash       string `json:"hash"`
	HashFormat string `json:"hash-format"`
}
