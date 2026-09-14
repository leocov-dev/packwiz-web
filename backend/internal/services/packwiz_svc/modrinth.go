package packwiz_svc

import (
	"codeberg.org/jmansfield/go-modrinth/modrinth"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/leocov-dev/packwiz-nxt/core"
	"github.com/leocov-dev/packwiz-nxt/sources"
	"packwiz-web/internal/log"
	"packwiz-web/internal/tables"
	"packwiz-web/internal/types/dto"
	"packwiz-web/internal/types/response"
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

	return mainMod, missingDependencies, err
}

// strPtr
// nil-safe dereference for the modrinth client's *string fields
func strPtr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func getInstalledModIdentifiers(mods []tables.Mod) (slugs map[string]bool, projectIDs map[string]bool) {
	slugs = make(map[string]bool)
	projectIDs = make(map[string]bool)

	addID := func(v any) {
		if v == nil {
			return
		}
		switch id := v.(type) {
		case string:
			if id != "" {
				projectIDs[id] = true
				slugs[strings.ToLower(id)] = true
			}
		case int:
			if id != 0 {
				s := strconv.Itoa(id)
				projectIDs[s] = true
				slugs[strings.ToLower(s)] = true
			}
		case int64:
			if id != 0 {
				s := strconv.FormatInt(id, 10)
				projectIDs[s] = true
				slugs[strings.ToLower(s)] = true
			}
		case float64:
			if id != 0 {
				s := strconv.Itoa(int(id))
				projectIDs[s] = true
				slugs[strings.ToLower(s)] = true
			}
		case uint:
			if id != 0 {
				s := strconv.FormatUint(uint64(id), 10)
				projectIDs[s] = true
				slugs[strings.ToLower(s)] = true
			}
		case uint32:
			if id != 0 {
				s := strconv.FormatUint(uint64(id), 10)
				projectIDs[s] = true
				slugs[strings.ToLower(s)] = true
			}
		}
	}

	for _, m := range mods {
		if m.Slug != "" {
			slugs[strings.ToLower(m.Slug)] = true
		}
		addID(m.Update["mod-id"])
		addID(m.Update["project-id"])
		if mr, ok := m.Update["modrinth"].(map[string]any); ok {
			addID(mr["mod-id"])
		} else if mr, ok := m.Update["modrinth"].(map[string]interface{}); ok {
			addID(mr["mod-id"])
		}
		if cf, ok := m.Update["curseforge"].(map[string]any); ok {
			addID(cf["project-id"])
		} else if cf, ok := m.Update["curseforge"].(map[string]interface{}); ok {
			addID(cf["project-id"])
		}
	}
	return slugs, projectIDs
}

// SearchModrinthProjects
// searches Modrinth by name/keyword, optionally filtered to compatible game versions,
// for the "search for a mod" picker in the add-mod flow. Returns up to 25 results and marks
// installed mods.
func (ps *PackwizService) SearchModrinthProjects(packId uint, query string, versions []string) ([]dto.ModSearchResult, response.ServerError) {
	var facets [][]string
	if len(versions) > 0 {
		versionFacets := make([]string, 0, len(versions))
		for _, v := range versions {
			versionFacets = append(versionFacets, "versions:"+v)
		}
		facets = append(facets, versionFacets)
	}

	searchRes, err := sources.GetModrinthClient().Projects.Search(&modrinth.SearchOptions{
		Limit:  25,
		Index:  "downloads",
		Query:  query,
		Facets: facets,
	})
	if err != nil {
		if err.Error() == "no projects found" {
			return []dto.ModSearchResult{}, nil
		}
		return nil, response.Wrap(err)
	}
	if searchRes == nil || len(searchRes.Hits) == 0 {
		return []dto.ModSearchResult{}, nil
	}

	var mods []tables.Mod
	if err := ps.db.Where("pack_id = ?", packId).Find(&mods).Error; err != nil {
		log.Warn(fmt.Sprintf("failed to get installed mods for pack %d: %v", packId, err))
	}
	slugs, projectIDs := getInstalledModIdentifiers(mods)

	results := make([]dto.ModSearchResult, 0, len(searchRes.Hits))
	for _, hit := range searchRes.Hits {
		slug := strPtr(hit.Slug)
		projId := strPtr(hit.ProjectID)
		isInstalled := (slug != "" && slugs[strings.ToLower(slug)]) ||
			(projId != "" && projectIDs[projId]) ||
			(projId != "" && slugs[strings.ToLower(projId)])

		results = append(results, dto.ModSearchResult{
			Slug:        slug,
			Title:       strPtr(hit.Title),
			Description: strPtr(hit.Description),
			IconUrl:     strPtr(hit.IconURL),
			ProjectId:   projId,
			Installed:   isInstalled,
		})
	}

	return results, nil
}
