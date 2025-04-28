package packwiz_schema

import (
	"packwiz-web/internal/types"
)

type IndexFile struct {
	HashFormat string      `toml:"hash-format"`
	Files      []IndexMeta `toml:"files"`
}

type ModFile struct {
	Name     string          `toml:"name"`
	Filename string          `toml:"filename"`
	Side     types.ModSide   `toml:"side"`
	Pin      bool            `toml:"pin,omitempty"`
	Download DownloadMeta    `toml:"download"`
	Update   UpdateSourceMap `toml:"update"`
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
	Modrinth   ModrinthMeta   `toml:"modrinth,omitempty"`
	Curseforge CurseforgeMeta `toml:"curseforge,omitempty"`
}

type DownloadMeta struct {
	Url        string `toml:"url,omitempty"`
	Mode       string `toml:"mode,omitempty"`
	HashFormat string `toml:"hash-format"`
	Hash       string `toml:"hash"`
}
