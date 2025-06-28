package packwiz_svc

import (
	"errors"
	"fmt"
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

	return mod, missingDependencies, nil
}
