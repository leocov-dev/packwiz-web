package types

type LoaderData struct {
	Type    string `json:"type"`
	Version string `json:"version"`
}

type VersionsData struct {
	Minecraft string     `json:"minecraft"`
	Loader    LoaderData `json:"loader"`
}

type PackData struct {
	Name       string       `json:"name"`
	Version    string       `json:"version"`
	PackFormat string       `json:"pack-format"`
	Versions   VersionsData `json:"versions"`
}
