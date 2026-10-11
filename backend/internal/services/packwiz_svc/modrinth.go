package packwiz_svc

import (
	"codeberg.org/jmansfield/go-modrinth/modrinth"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/leocov-dev/packwiz-nxt/core"
	"github.com/leocov-dev/packwiz-nxt/sources"
	"packwiz-web/internal/log"
	"packwiz-web/internal/tables"
	"packwiz-web/internal/types/dto"
	"packwiz-web/internal/types/response"
)

// errNoCompatibleVersion marks a mod that exists but has no version for the
// pack's Minecraft version(s) and loader.
var errNoCompatibleVersion = errors.New("no compatible version")

func lookupModrinthDependencies(url string, pack core.Pack) ([]*core.Mod, error) {
	_, version, err := modrinthProjectAndVersion(url, pack)
	if err != nil {
		return nil, err
	}
	return modrinthMissingDependencies(version, pack)
}

func modrinthMissingDependencies(version *modrinth.Version, pack core.Pack) ([]*core.Mod, error) {
	if len(version.Dependencies) == 0 {
		return nil, nil
	}
	return sources.ModrinthFindMissingDependencies(version, pack, "")
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
		// the project is returned so callers can name it
		return project, nil, fmt.Errorf("version lookup failed: %w: %w", errNoCompatibleVersion, err)
	}
	if version == nil {
		return project, nil, fmt.Errorf("version not found for project %s: %w", *project.ID, errNoCompatibleVersion)
	}
	log.Debug("version: ", *version.ID, *version.Name)

	return project, version, nil
}

// resolveModrinthMod resolves a Modrinth URL or slug to the mod to add, its
// missing dependencies and the projects its version declares incompatible.
func resolveModrinthMod(url string, pack core.Pack) (resolvedMod, error) {
	project, version, err := modrinthProjectAndVersion(url, pack)
	if err != nil {
		if project != nil {
			return resolvedMod{Ref: modrinthProjectRef(project)}, err
		}
		return resolvedMod{}, err
	}

	mainMod, err := sources.ModrinthNewMod(project, version, "", pack.GetCompatibleLoaders(), "")
	if err != nil {
		return resolvedMod{}, err
	}
	if mainMod == nil {
		return resolvedMod{}, errors.New("failed to add mod")
	}

	deps, err := modrinthMissingDependencies(version, pack)
	if err != nil {
		return resolvedMod{}, err
	}

	return resolvedMod{
		Ref:          modrinthProjectRef(project),
		Mod:          mainMod,
		Deps:         deps,
		Incompatible: modrinthIncompatibleProjects(version),
	}, nil
}

func modrinthProjectRef(project *modrinth.Project) projectRef {
	return projectRef{
		Source:    sourceModrinth,
		ProjectID: strPtr(project.ID),
		Slug:      strPtr(project.Slug),
		Name:      strPtr(project.Title),
	}
}

// modrinthIncompatibleProjects lists the project ids a version declares
// incompatible. Declarations that only name a version id are ignored.
func modrinthIncompatibleProjects(version *modrinth.Version) []string {
	var out []string
	for _, dep := range version.Dependencies {
		if dep == nil || strPtr(dep.DependencyType) != "incompatible" || strPtr(dep.ProjectID) == "" {
			continue
		}
		out = append(out, *dep.ProjectID)
	}
	return out
}

func addModrinthMod(url string, pack core.Pack) (*core.Mod, []*core.Mod, error) {
	resolved, err := resolveModrinthMod(url, pack)
	if err != nil {
		return nil, nil, err
	}
	return resolved.Mod, resolved.Deps, nil
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
	var pack tables.Pack
	if err := ps.db.Where("id = ?", packId).First(&pack).Error; err != nil {
		return nil, response.New(http.StatusNotFound, "Pack not found")
	}

	hits, err := searchModrinthHits(pack.AsMeta(), query, versions, 25)
	if err != nil {
		return nil, response.Wrap(err)
	}
	if len(hits) == 0 {
		return []dto.ModSearchResult{}, nil
	}

	var mods []tables.Mod
	if err := ps.db.Where("pack_id = ?", packId).Find(&mods).Error; err != nil {
		log.Warn(fmt.Sprintf("failed to get installed mods for pack %d: %v", packId, err))
	}
	slugs, projectIDs := getInstalledModIdentifiers(mods)

	results := make([]dto.ModSearchResult, 0, len(hits))
	for _, hit := range hits {
		slug := strPtr(hit.Slug)
		projId := strPtr(hit.ProjectID)
		isInstalled := (slug != "" && slugs[strings.ToLower(slug)]) ||
			(projId != "" && projectIDs[projId]) ||
			(projId != "" && slugs[strings.ToLower(projId)])

		results = append(results, modrinthResultFromHit(hit, isInstalled))
	}

	return results, nil
}

// searchModrinthHits searches Modrinth mods compatible with the pack's
// loaders, and with versions when given, most downloaded first.
func searchModrinthHits(pack core.Pack, query string, versions []string, limit int) ([]*modrinth.SearchResult, error) {
	var facets [][]string
	if len(versions) > 0 {
		versionFacets := make([]string, 0, len(versions))
		for _, v := range versions {
			versionFacets = append(versionFacets, "versions:"+v)
		}
		facets = append(facets, versionFacets)
	}

	if loaders := pack.GetCompatibleLoaders(); len(loaders) > 0 {
		loaderFacets := make([]string, 0, len(loaders))
		for _, l := range loaders {
			loaderFacets = append(loaderFacets, "categories:"+l)
		}
		facets = append(facets, loaderFacets)
	}

	searchRes, err := sources.GetModrinthClient().Projects.Search(&modrinth.SearchOptions{
		Limit:  limit,
		Index:  "downloads",
		Query:  query,
		Facets: facets,
	})
	if err != nil {
		if err.Error() == "no projects found" {
			return nil, nil
		}
		return nil, err
	}
	if searchRes == nil {
		return nil, nil
	}
	return searchRes.Hits, nil
}
