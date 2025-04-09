package types

type ModrinthData struct {
	ModrinthId string `json:"modId"`
	Version    string `json:"version"`
}

type CurseforgeData struct {
	FileId    int `json:"fileId"`
	ProjectId int `json:"projectId"`
}

type ModSource struct {
	// modrinth or curseforge
	Type string `json:"type"`
	// modrinth info
	ModId   string `json:"modId,omitempty"`
	Version string `json:"version,omitempty"`
	// curseforge info
	FileId    int `json:"fileId,omitempty"`
	ProjectId int `json:"projectId,omitempty"`
}

type ModSide string

const (
	SideClient ModSide = "client"
	SideServer ModSide = "server"
	SideBoth   ModSide = "both"
)

type ModData struct {
	Name        string    `json:"name"`
	DisplayName string    `json:"displayName"`
	Type        string    `json:"type"`
	Filename    string    `json:"filename"`
	Side        ModSide   `json:"side"`
	Pinned      bool      `json:"pinned"`
	Source      ModSource `json:"source"`
	SourceLink  string    `json:"sourceLink"`
}
