package packwiz_svc

import (
	"strconv"
	"strings"

	"codeberg.org/jmansfield/go-modrinth/modrinth"
	"github.com/leocov-dev/packwiz-nxt/sources"

	"packwiz-web/internal/types/dto"
)

// knownLoaders is the set of Modrinth "categories" that are really loaders/platforms.
var knownLoaders = map[string]struct{}{
	"fabric": {}, "forge": {}, "neoforge": {}, "quilt": {}, "liteloader": {},
	"rift": {}, "modloader": {}, "bukkit": {}, "spigot": {}, "paper": {},
	"purpur": {}, "folia": {}, "sponge": {}, "velocity": {}, "bungeecord": {},
	"waterfall": {}, "datapack": {}, "minecraft": {}, "iris": {}, "optifine": {},
	"canvas": {}, "vanilla": {}, "babric": {}, "bta-babric": {}, "legacy-fabric": {},
	"ornithe": {}, "nilloader": {}, "java-agent": {}, "geyser": {}, "leaf": {},
}

// cfLoaderNames maps the CurseForge modLoader enum to a loader name.
// Cauldron and Any are intentionally omitted.
var cfLoaderNames = map[sources.ModloaderType]string{
	sources.ModloaderTypeForge:      "forge",
	sources.ModloaderTypeLiteloader: "liteloader",
	sources.ModloaderTypeFabric:     "fabric",
	sources.ModloaderTypeQuilt:      "quilt",
	sources.ModloaderTypeNeoForge:   "neoforge",
}

// splitModrinthCategories separates Modrinth categories into loaders and
// regular categories, preserving order and dropping duplicates.
func splitModrinthCategories(cats []string) (loaders, categories []string) {
	seen := map[string]struct{}{}
	for _, c := range cats {
		c = strings.ToLower(strings.TrimSpace(c))
		if c == "" {
			continue
		}
		if _, dup := seen[c]; dup {
			continue
		}
		seen[c] = struct{}{}
		if _, ok := knownLoaders[c]; ok {
			loaders = append(loaders, c)
		} else {
			categories = append(categories, c)
		}
	}
	return loaders, categories
}

func downloadsPtr(d *uint32) uint64 {
	if d == nil {
		return 0
	}
	return uint64(*d)
}

func cfAuthor(item cfSearchItem) string {
	for _, a := range item.Authors {
		if n := strings.TrimSpace(a.Name); n != "" {
			return n
		}
	}
	return ""
}

func cfDownloads(item cfSearchItem) uint64 {
	if item.DownloadCount == "" {
		return 0
	}
	if u, err := strconv.ParseUint(item.DownloadCount.String(), 10, 64); err == nil {
		return u
	}
	if f, err := item.DownloadCount.Float64(); err == nil && f > 0 {
		return uint64(f)
	}
	return 0
}

// cfLoaders returns the deduplicated loader names in first-seen order.
// Note: latestFilesIndexes only covers the latest file per game version (and
// loader), so without a gameVersion filter the result may reflect loaders from
// old Minecraft versions rather than the current ones.
func cfLoaders(item cfSearchItem) []string {
	var out []string
	seen := map[string]struct{}{}
	for _, f := range item.LatestFilesIndexes {
		name, ok := cfLoaderNames[sources.ModloaderType(f.ModLoader)]
		if !ok {
			continue
		}
		if _, dup := seen[name]; dup {
			continue
		}
		seen[name] = struct{}{}
		out = append(out, name)
	}
	return out
}

func cfCategories(item cfSearchItem) []string {
	var out []string
	seen := map[string]struct{}{}
	for _, c := range item.Categories {
		n := strings.TrimSpace(c.Name)
		if n == "" {
			continue
		}
		if _, dup := seen[n]; dup {
			continue
		}
		seen[n] = struct{}{}
		out = append(out, n)
	}
	return out
}

// modrinthResultFromHit maps a Modrinth search hit to the search DTO.
func modrinthResultFromHit(hit *modrinth.SearchResult, installed bool) dto.ModSearchResult {
	loaders, categories := splitModrinthCategories(hit.Categories)
	return dto.ModSearchResult{
		Author:      strPtr(hit.Author),
		Downloads:   downloadsPtr(hit.Downloads),
		Loaders:     loaders,
		Categories:  categories,
		Slug:        strPtr(hit.Slug),
		Title:       strPtr(hit.Title),
		Description: strPtr(hit.Description),
		IconUrl:     strPtr(hit.IconURL),
		ProjectId:   strPtr(hit.ProjectID),
		Installed:   installed,
	}
}

// curseforgeResultFromItem maps a CurseForge search item to the search DTO.
func curseforgeResultFromItem(item cfSearchItem, projId, iconUrl string, installed bool) dto.ModSearchResult {
	return dto.ModSearchResult{
		Author:      cfAuthor(item),
		Downloads:   cfDownloads(item),
		Loaders:     cfLoaders(item),
		Categories:  cfCategories(item),
		Slug:        item.Slug,
		Title:       item.Name,
		Description: item.Summary,
		IconUrl:     iconUrl,
		ProjectId:   projId,
		Installed:   installed,
	}
}
