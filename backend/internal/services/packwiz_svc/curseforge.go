package packwiz_svc

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/leocov-dev/packwiz-nxt/core"
	"github.com/leocov-dev/packwiz-nxt/sources"
)

func lookupCurseforgeDependencies(url string, pack core.Pack) ([]*core.Mod, error) {
	_, fileInfoData, err := curseforgeModInfoFromUrl(url, pack)
	if err != nil {
		return nil, fmt.Errorf("failed to get mod info: %w", err)
	}

	primaryMCVersion, err := pack.GetMCVersion()
	if err != nil {
		return nil, fmt.Errorf("failed to get primary MC version: %w", err)
	}

	if len(fileInfoData.Dependencies) > 0 {
		return sources.CurseforgeFindMissingDependencies(pack, *fileInfoData, primaryMCVersion)
	}

	// no dependencies
	return nil, nil
}

func curseforgeModInfoFromUrl(url string, pack core.Pack) (*sources.CfModInfo, *sources.CfModFileInfo, error) {
	mcVersions, err := pack.GetSupportedMCVersions()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get supported MC versions: %w", err)
	}

	category, slug, fileID, err := sources.CurseforgeParseUrl(url)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse Curseforge URL: %w", err)
	}

	modInfo, fileInfo, err := sources.CurseforgeModInfoFromSlug(
		slug,
		category,
		fileID,
		mcVersions,
		sources.CfGetSearchLoaderType(pack),
		pack.GetCompatibleLoaders(),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get mod info: %w", err)
	}
	// nxt returns zero values and no error when the slug matches nothing
	if modInfo.ID == 0 {
		return nil, nil, fmt.Errorf("curseforge project not found: %s", slug)
	}

	return &modInfo, &fileInfo, nil
}

func addCurseforgeMod(url string, pack core.Pack) (*core.Mod, []*core.Mod, error) {

	modInfoData, fileInfoData, err := curseforgeModInfoFromUrl(url, pack)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get mod info: %w", err)
	}

	mod, err := sources.CurseforgeNewMod(*modInfoData, *fileInfoData, false)
	if err != nil {
		return nil, nil, err
	}

	if mod == nil {
		return nil, nil, errors.New("failed to add mod")
	}

	missingDependencies, err := lookupCurseforgeDependencies(url, pack)

	return mod, missingDependencies, err
}

// curseforgeIncompatible is CurseForge's "incompatible" file relation type.
const curseforgeIncompatible = 5

// resolveCurseforgeByID resolves a CurseForge project id to the mod to add,
// its missing dependencies and the projects its file declares incompatible.
func resolveCurseforgeByID(modID uint32, pack core.Pack) (resolvedMod, error) {
	mcVersions, err := pack.GetSupportedMCVersions()
	if err != nil {
		return resolvedMod{}, fmt.Errorf("failed to get supported MC versions: %w", err)
	}
	primaryMCVersion, err := pack.GetMCVersion()
	if err != nil {
		return resolvedMod{}, fmt.Errorf("failed to get primary MC version: %w", err)
	}

	modInfo, err := sources.GetCurseforgeClient().GetModInfo(modID)
	if err != nil {
		return resolvedMod{}, fmt.Errorf("failed to get mod info: %w", err)
	}
	if modInfo.ID == 0 {
		return resolvedMod{}, fmt.Errorf("curseforge project not found: %d", modID)
	}

	ref := projectRef{
		Source:    sourceCurseforge,
		ProjectID: strconv.FormatUint(uint64(modInfo.ID), 10),
		Slug:      modInfo.Slug,
		Name:      modInfo.Name,
	}

	fileInfo, err := sources.GetLatestFile(modInfo, mcVersions, 0, pack.GetCompatibleLoaders())
	if err != nil {
		return resolvedMod{Ref: ref}, fmt.Errorf("file lookup failed: %w: %w", errNoCompatibleVersion, err)
	}

	mod, err := sources.CurseforgeNewMod(modInfo, fileInfo, false)
	if err != nil {
		return resolvedMod{}, err
	}
	if mod == nil {
		return resolvedMod{}, errors.New("failed to add mod")
	}

	var deps []*core.Mod
	var incompatible []string
	for _, dep := range fileInfo.Dependencies {
		if uint8(dep.Type) == curseforgeIncompatible {
			incompatible = append(incompatible, strconv.FormatUint(uint64(dep.ModID), 10))
		}
	}
	if len(fileInfo.Dependencies) > 0 {
		deps, err = sources.CurseforgeFindMissingDependencies(pack, fileInfo, primaryMCVersion)
		if err != nil {
			return resolvedMod{}, err
		}
	}

	return resolvedMod{
		Ref:          ref,
		Mod:          mod,
		Deps:         deps,
		Incompatible: incompatible,
	}, nil
}
