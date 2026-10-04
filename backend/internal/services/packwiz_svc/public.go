package packwiz_svc

import (
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"packwiz-web/internal/params"
	"packwiz-web/internal/tables"
	"packwiz-web/internal/types"
	"packwiz-web/internal/types/dto"
	"packwiz-web/internal/types/response"
)

// GetPublicPack returns the unauthenticated view of a pack. Only packs that are
// public, published and not archived are visible; anything else is a 404 so
// the response does not reveal that a private pack exists.
func (ps *PackwizService) GetPublicPack(slug, scheme, host string) (dto.PublicPackResponse, response.ServerError) {
	var pack tables.Pack

	err := ps.db.
		Preload("Author").
		Preload("Mods").
		Where("slug = ? AND is_public = ? AND status = ?", slug, true, types.PackStatusPublished).
		First(&pack).Error
	if err != nil {
		return dto.PublicPackResponse{}, response.New(http.StatusNotFound, "pack not found")
	}

	resp := buildPublicPack(pack, packTomlURL(scheme, host, publicLinkKey, pack.Slug))
	resp.MultiMCURL = instanceZipURL(scheme, host, publicLinkKey, pack)
	return resp, nil
}

// publicLinkKey is the token segment used in links to public packs.
const publicLinkKey = "public"

func packTomlURL(scheme, host, key, slug string) string {
	return packFileURL(scheme, host, key, slug, "pack.toml")
}

// instanceZipURL is the MultiMC / Prism instance zip next to pack.toml. It is
// authenticated by the same key, so a launcher can import it by URL. The last
// segment is the pack name: Prism suggests the instance name from it.
func instanceZipURL(scheme, host, key string, pack tables.Pack) string {
	return packFileURL(scheme, host, key, pack.Slug, params.InstanceZipDir+"/"+InstanceZipFileName(pack.Name, pack.Slug))
}

// InstanceZipFileName is the zip's file name: the pack name without characters
// that are unsafe in a file name on any OS, falling back to the slug. Prism and
// MultiMC suggest the imported instance's name from it.
func InstanceZipFileName(packName, slug string) string {
	name := strings.Map(func(r rune) rune {
		if r < 0x20 || r == 0x7f || strings.ContainsRune(`<>:"/\|?*`, r) {
			return -1
		}
		return r
	}, packName)
	name = strings.Join(strings.Fields(name), " ")
	name = strings.Trim(name, " .")
	if name == "" {
		name = slug
	}
	return name + ".zip"
}

func packFileURL(scheme, host, key, slug, file string) string {
	return (&url.URL{
		Scheme: scheme,
		Host:   host,
		Path:   fmt.Sprintf("/packwiz/%s/%s/%s", key, slug, file),
	}).String()
}

// linkKey is the token segment of a consumer link: "public" for a public pack,
// otherwise the user's own link token.
func linkKey(pack tables.Pack, userToken string) string {
	if pack.IsPublic {
		return publicLinkKey
	}
	return userToken
}

// ConsumerPackTomlLink is the pack.toml link for a consumer request that has
// already been authenticated with token.
func (ps *PackwizService) ConsumerPackTomlLink(pack tables.Pack, token, scheme, host string) string {
	return packTomlURL(scheme, host, linkKey(pack, token), pack.Slug)
}

// buildPublicPack maps a pack to its public view. It is pure so it can be tested
// without a database.
func buildPublicPack(pack tables.Pack, tomlURL string) dto.PublicPackResponse {
	mods := make([]dto.PublicPackMod, 0, len(pack.Mods))
	for _, m := range pack.Mods {
		mods = append(mods, dto.PublicPackMod{
			Slug:         m.Slug,
			Name:         m.Name,
			Type:         m.Type,
			Side:         string(m.Side),
			Version:      m.Version,
			Source:       m.Source,
			Optional:     m.Option.Optional,
			IsDependency: m.IsDependency,
		})
	}
	sort.SliceStable(mods, func(i, j int) bool {
		return strings.ToLower(mods[i].Name) < strings.ToLower(mods[j].Name)
	})

	acceptable := []string(pack.AcceptableGameVersions)
	if acceptable == nil {
		acceptable = []string{}
	}

	return dto.PublicPackResponse{
		Slug:                   pack.Slug,
		Name:                   pack.Name,
		Description:            pack.Description,
		Author:                 pack.Author.Username,
		Version:                pack.Version,
		MCVersion:              pack.MCVersion,
		Loader:                 pack.Loader,
		LoaderVersion:          pack.LoaderVersion,
		AcceptableGameVersions: acceptable,
		PackFormat:             pack.PackFormat,
		UpdatedAt:              pack.UpdatedAt,
		PackTomlURL:            tomlURL,
		Mods:                   mods,
	}
}
