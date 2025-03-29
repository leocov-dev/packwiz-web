package loaders

import (
	"encoding/json"
	"io"
	"net/http"
	"packwiz-web/internal/types/response"
)

type VersionInfo struct {
	Latest struct {
		Release  string `json:"release"`
		Snapshot string `json:"snapshot"`
	} `json:"latest"`
	Versions []Version `json:"versions"`
}

type Version struct {
	ID          string `json:"id"`
	Type        string `json:"type"`
	URL         string `json:"url"`
	Time        string `json:"time"`
	ReleaseTime string `json:"releaseTime"`
}

type McVersionInfo struct {
	Latest         string   `json:"latest"`
	LatestSnapshot string   `json:"snapshot"`
	Versions       []string `json:"versions"`
}

func GetMinecraftVersions() (McVersionInfo, response.ServerError) {
	var versionInfo McVersionInfo

	resp, err := http.Get("https://launchermeta.mojang.com/mc/game/version_manifest.json")
	if err != nil {
		return versionInfo, response.Wrap(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return versionInfo, response.Wrap(err)
	}

	var info VersionInfo
	if err := json.Unmarshal(body, &info); err != nil {
		return versionInfo, response.Wrap(err)
	}

	versions := make([]string, 0)

	for _, v := range info.Versions {
		if v.Type != "release" {
			continue
		}
		versions = append(versions, v.ID)
	}

	versionInfo = McVersionInfo{
		Latest:         info.Latest.Release,
		LatestSnapshot: info.Latest.Snapshot,
		Versions:       versions,
	}

	return versionInfo, nil
}
