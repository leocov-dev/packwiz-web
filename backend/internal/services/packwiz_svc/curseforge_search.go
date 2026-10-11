package packwiz_svc

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	libConfig "github.com/leocov-dev/packwiz-nxt/config"
	"github.com/leocov-dev/packwiz-nxt/core"
	"github.com/leocov-dev/packwiz-nxt/sources"
	"packwiz-web/internal/config"
	"packwiz-web/internal/log"
	"packwiz-web/internal/tables"
	"packwiz-web/internal/types/dto"
	"packwiz-web/internal/types/response"
)

type cfSearchItem struct {
	ID      uint32 `json:"id"`
	Name    string `json:"name"`
	Slug    string `json:"slug"`
	Summary string `json:"summary"`
	Logo    struct {
		ThumbnailURL string `json:"thumbnailUrl"`
		URL          string `json:"url"`
	} `json:"logo"`
	Authors []struct {
		Name string `json:"name"`
	} `json:"authors"`
	// DownloadCount is a json.Number so an integer (documented) or a float
	// (e.g. 1.0e6) both decode; a strict uint64 would fail the whole search.
	DownloadCount      json.Number `json:"downloadCount"`
	LatestFilesIndexes []struct {
		ModLoader int `json:"modLoader"`
	} `json:"latestFilesIndexes"`
	Categories []struct {
		Name string `json:"name"`
	} `json:"categories"`
}

type cfSearchResponse struct {
	Data []cfSearchItem `json:"data"`
}

// SearchCurseforgeProjects searches CurseForge by name/keyword for a pack.
// Returns up to 25 results and marks already installed mods.
func (ps *PackwizService) SearchCurseforgeProjects(packId uint, query string, versions []string) ([]dto.ModSearchResult, response.ServerError) {
	if !config.HasCurseforgeApiKey() {
		return nil, response.New(http.StatusBadRequest, "CurseForge API key not configured")
	}

	var pack tables.Pack
	if err := ps.db.Where("id = ?", packId).First(&pack).Error; err != nil {
		return nil, response.New(http.StatusNotFound, "Pack not found")
	}

	searchRes, err := searchCurseforgeItems(pack, query, versions, 25)
	if err != nil {
		if errors.Is(err, errCurseforgeKey) {
			return nil, response.New(http.StatusBadRequest, err.Error())
		}
		return nil, response.Wrap(err)
	}

	var mods []tables.Mod
	if err := ps.db.Where("pack_id = ?", packId).Find(&mods).Error; err != nil {
		log.Warn(fmt.Sprintf("failed to get installed mods for pack %d: %v", packId, err))
	}
	slugs, projectIDs := getInstalledModIdentifiers(mods)

	results := make([]dto.ModSearchResult, 0, len(searchRes))
	for _, item := range searchRes {
		slug := item.Slug
		projId := strconv.FormatUint(uint64(item.ID), 10)

		iconUrl := item.Logo.ThumbnailURL
		if iconUrl == "" {
			iconUrl = item.Logo.URL
		}

		isInstalled := (slug != "" && slugs[strings.ToLower(slug)]) ||
			(projId != "" && projectIDs[projId]) ||
			(projId != "" && slugs[strings.ToLower(projId)])

		results = append(results, curseforgeResultFromItem(item, projId, iconUrl, isInstalled))
	}

	return results, nil
}

// errCurseforgeKey marks a configured CurseForge API key that cannot be decoded.
var errCurseforgeKey = errors.New("invalid CurseForge API key")

// searchCurseforgeItems searches CurseForge mods compatible with the pack's
// Minecraft version (or versions[0]) and loader, most popular first. Callers
// check config.HasCurseforgeApiKey first.
func searchCurseforgeItems(pack tables.Pack, query string, versions []string, pageSize int) ([]cfSearchItem, error) {
	apiKey, err := libConfig.DecodeCfApiKey()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errCurseforgeKey, err)
	}

	params := url.Values{}
	params.Set("gameId", "432") // Minecraft
	params.Set("classId", "6")  // Minecraft Mods
	params.Set("pageSize", strconv.Itoa(pageSize))
	params.Set("sortField", "2") // Sort by popularity
	params.Set("sortOrder", "desc")

	if query != "" {
		params.Set("searchFilter", query)
	}

	// Determine Minecraft version
	gameVersion := ""
	if len(versions) > 0 && versions[0] != "" {
		gameVersion = versions[0]
	} else if pack.MCVersion != "" {
		gameVersion = pack.MCVersion
	}
	if gameVersion != "" {
		params.Set("gameVersion", gameVersion)
	}

	// Determine loader
	metaPack := pack.AsMeta()
	loaderType := sources.CfGetSearchLoaderType(metaPack)
	if loaderType == sources.ModloaderTypeAny {
		switch strings.ToLower(pack.Loader) {
		case "fabric":
			loaderType = sources.ModloaderTypeFabric
		case "forge":
			loaderType = sources.ModloaderTypeForge
		case "neoforge":
			loaderType = sources.ModloaderTypeNeoForge
		}
	}
	if loaderType != sources.ModloaderTypeAny {
		params.Set("modLoaderType", strconv.Itoa(int(loaderType)))
	}

	reqUrl := "https://api.curseforge.com/v1/mods/search?" + params.Encode()
	req, err := http.NewRequest("GET", reqUrl, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create CurseForge request: %w", err)
	}

	req.Header.Set("User-Agent", core.UserAgent)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("x-api-key", apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("CurseForge search request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("CurseForge search returned status %d: %s", resp.StatusCode, string(body))
	}

	var searchRes cfSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchRes); err != nil && err != io.EOF {
		return nil, fmt.Errorf("failed to parse CurseForge search response: %w", err)
	}

	return searchRes.Data, nil
}
