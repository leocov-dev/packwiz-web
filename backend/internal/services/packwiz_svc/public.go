package packwiz_svc

import (
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"

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

	return buildPublicPack(pack, packTomlURL(scheme, host, publicLinkKey, pack.Slug)), nil
}

// publicLinkKey is the token segment used in links to public packs.
const publicLinkKey = "public"

func packTomlURL(scheme, host, key, slug string) string {
	return (&url.URL{
		Scheme: scheme,
		Host:   host,
		Path:   fmt.Sprintf("/packwiz/%s/%s/pack.toml", key, slug),
	}).String()
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
