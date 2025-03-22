package loaders

import (
	"encoding/xml"
	"io"
	"net/http"
	"packwiz-web/internal/utils"
	"strings"
)

// VersionMap represents a map of Minecraft versions to their corresponding Forge versions
type VersionMap map[string][]string

type Loaders struct {
	Fabric     []string   `json:"fabric"`
	Forge      VersionMap `json:"forge"`
	Liteloader []string   `json:"liteloader"`
	Quilt      []string   `json:"quilt"`
	Neoforge   VersionMap `json:"neoforge"`
}

type mavenXmlMetadata struct {
	Versioning struct {
		Versions struct {
			Version []string `xml:"version"`
		} `xml:"versions"`
	} `xml:"versioning"`
}

func GetLoadersAndVersions() (Loaders, error) {
	var LoaderVersions Loaders

	if fabricVersions, err := fetchFabricVersions(); err != nil {
		return LoaderVersions, err
	} else {
		LoaderVersions.Fabric = fabricVersions
	}

	if forgeVersions, err := fetchForgeVersions(); err != nil {
		return LoaderVersions, err
	} else {
		LoaderVersions.Forge = forgeVersions
	}

	if liteloaderVersions, err := fetchLiteloaderVersions(); err != nil {
		return LoaderVersions, err
	} else {
		LoaderVersions.Liteloader = liteloaderVersions
	}

	if quiltVersions, err := fetchQuiltVersions(); err != nil {
		return LoaderVersions, err
	} else {
		LoaderVersions.Quilt = quiltVersions
	}

	if neoforgeVersions, err := fetchNeoforgeVersions(); err != nil {
		return LoaderVersions, err
	} else {
		LoaderVersions.Neoforge = neoforgeVersions
	}

	return LoaderVersions, nil
}

func fetchFabricVersions() ([]string, error) {
	versions, err := fetchMavenList(
		"https://maven.fabricmc.net/net/fabricmc/fabric-loader/maven-metadata.xml",
		func(version string) string {
			// Skip versions containing "+"
			if strings.Contains(version, "+") {
				return ""
			}
			return version
		},
	)

	return utils.SortDescending(versions), err
}

func fetchForgeVersions() (VersionMap, error) {
	versionMap, err := fetchMavenMap(
		"https://maven.minecraftforge.net/net/minecraftforge/forge/maven-metadata.xml",
		func(version string) (string, string) {
			parts := strings.Split(version, "-")

			return parts[0], parts[1]
		},
	)

	for mcVersion, loaderVersions := range versionMap {
		versionMap[mcVersion] = utils.SortDescending(loaderVersions)
	}

	return versionMap, err
}

func fetchLiteloaderVersions() ([]string, error) {
	versions, err := fetchMavenList(
		"https://repo.mumfrey.com/content/repositories/snapshots/com/mumfrey/liteloader/maven-metadata.xml",
		func(version string) string {
			// versions are in the format <version>-SNAPSHOT
			return strings.Split(version, "-")[0]
		},
	)
	return utils.SortDescending(versions), err
}

func fetchQuiltVersions() ([]string, error) {
	versions, err := fetchMavenList(
		"https://maven.quiltmc.org/repository/release/org/quiltmc/quilt-loader/maven-metadata.xml",
		func(version string) string {
			return version
		},
	)

	return utils.SortDescending(versions), err
}

func fetchNeoforgeVersions() (VersionMap, error) {
	versions, err := fetchMavenMap(
		"https://maven.neoforged.net/releases/net/neoforged/forge/maven-metadata.xml",
		func(version string) (string, string) {
			parts := strings.Split(version, "-")
			if len(parts) < 2 {
				return "", ""
			}

			return parts[0], parts[1]
		},
	)
	if err != nil {
		return nil, err
	}

	moreVersions, err := fetchMavenMap(
		"https://maven.neoforged.net/releases/net/neoforged/neoforge/maven-metadata.xml",
		func(version string) (string, string) {
			parts := strings.Split(version, ".")

			if len(parts) < 2 {
				return "", ""
			}

			return "1." + parts[0] + "." + parts[1], version
		},
	)
	if err != nil {
		return nil, err
	}

	for mcVersion, loaderVersions := range moreVersions {
		if _, exists := versions[mcVersion]; !exists {
			versions[mcVersion] = make([]string, 0)
		}
		versions[mcVersion] = append(versions[mcVersion], loaderVersions...)
	}

	for mcVersion, loaderVersions := range versions {
		versions[mcVersion] = utils.SortDescending(loaderVersions)
	}

	return versions, nil
}

// ----
func fetchMavenList(url string, versionCb func(version string) string) ([]string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var metadata mavenXmlMetadata
	if err := xml.Unmarshal(body, &metadata); err != nil {
		return nil, err
	}

	var filteredVersions []string
	// we want in reverse order
	for i := len(metadata.Versioning.Versions.Version) - 1; i >= 0; i-- {
		version := metadata.Versioning.Versions.Version[i]

		processedVersion := versionCb(version)
		if processedVersion == "" {
			continue
		}

		filteredVersions = append(filteredVersions, processedVersion)
	}

	return filteredVersions, nil
}

func fetchMavenMap(url string, keyValueCb func(version string) (string, string)) (VersionMap, error) {

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var metadata mavenXmlMetadata
	if err := xml.Unmarshal(body, &metadata); err != nil {
		return nil, err
	}

	versionMap := make(VersionMap)

	for _, version := range metadata.Versioning.Versions.Version {
		if version == "" {
			continue
		}

		minecraftVersion, loaderVersion := keyValueCb(version)

		if minecraftVersion == "" || loaderVersion == "" {
			continue
		}

		if _, exists := versionMap[minecraftVersion]; !exists {
			versionMap[minecraftVersion] = make([]string, 0)
		}
		versionMap[minecraftVersion] = append(versionMap[minecraftVersion], loaderVersion)

	}

	return versionMap, nil
}
