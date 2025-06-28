package packwiz_svc

import (
	"codeberg.org/jmansfield/go-modrinth/modrinth"
	"errors"
	"fmt"
	"github.com/leocov-dev/packwiz-nxt/core"
	"github.com/leocov-dev/packwiz-nxt/sources"
	"packwiz-web/internal/log"
)

func lookupModrinthDependencies(url string, pack core.Pack) ([]*core.Mod, error) {
	var err error

	_, version, err := modrinthProjectAndVersion(url, pack)
	if err != nil {
		return nil, err
	}

	var missingDependencies []*core.Mod
	if len(version.Dependencies) > 0 {

		missingDependencies, err = sources.ModrinthFindMissingDependencies(version, pack, "")
		if err != nil {
			return nil, err
		}
	}

	return missingDependencies, nil
}

func modrinthProjectAndVersion(url string, pack core.Pack) (*modrinth.Project, *modrinth.Version, error) {
	projectSlug := sources.ParseAsModrinthSlug(url)
	if projectSlug == "" {
		return nil, nil, errors.New("invalid modrinth url")
	}
	log.Debug("project slug: ", projectSlug)

	project, err := sources.GetModrinthClient().Projects.Get(projectSlug)
	if err != nil {
		return nil, nil, fmt.Errorf("project lookup failed: %w", err)
	}
	if project == nil {
		return nil, nil, fmt.Errorf("project not found for slug: %s", projectSlug)
	}
	log.Debug("project: ", *project.ID, *project.Title)

	version, err := sources.ModrinthGetLatestVersion(*project.ID, *project.Title, pack, "")
	if err != nil {
		return nil, nil, fmt.Errorf("version lookup failed: %w", err)
	}
	if version == nil {
		return nil, nil, fmt.Errorf("version not found for project: %s", *project.ID)
	}
	log.Debug("version: ", *version.ID, *version.Name)

	return project, version, nil
}

func addModrinthMod(url string, pack core.Pack) (*core.Mod, []*core.Mod, error) {
	project, version, err := modrinthProjectAndVersion(url, pack)
	if err != nil {
		return nil, nil, err
	}

	mainMod, err := sources.ModrinthNewMod(project, version, "", pack.GetCompatibleLoaders(), "")
	if err != nil {
		return nil, nil, err
	}

	if mainMod == nil {
		return nil, nil, errors.New("failed to add mod")
	}

	missingDependencies, err := lookupModrinthDependencies(url, pack)

	return mainMod, missingDependencies, nil
}
