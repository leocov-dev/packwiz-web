package multimc_svc

import (
	"errors"
	"net/http"
	"net/url"

	"gorm.io/gorm"

	"packwiz-web/internal/services/packwiz_svc"
	"packwiz-web/internal/tables"
	"packwiz-web/internal/types/response"
)

// Archive is a generated instance zip and the file name to download it as.
type Archive struct {
	Filename string
	Data     []byte
}

// MultiMCService builds MultiMC / Prism Launcher instance archives for packs.
type MultiMCService struct {
	packwizSvc *packwiz_svc.PackwizService
	// publicURL is the configured external base URL, or nil if not set.
	publicURL *url.URL
}

// NewMultiMCService builds the service. publicURL is the app's external base
// URL (PWW_PUBLIC_URL); when set, the links inside an instance use it instead
// of the host the download came from.
func NewMultiMCService(db *gorm.DB, packwizSvc *packwiz_svc.PackwizService, publicURL string) *MultiMCService {
	return &MultiMCService{packwizSvc: packwizSvc, publicURL: parseBaseURL(publicURL)}
}

func parseBaseURL(raw string) *url.URL {
	u, err := url.Parse(raw)
	if raw == "" || err != nil || u.Scheme == "" || u.Host == "" {
		return nil
	}
	return u
}

// origin picks the scheme and host for links baked into an instance. The zip
// lives on in the user's launcher, so a configured public URL wins over the
// request's host (which may be a LAN address, or a spoofed Host header).
func (ms *MultiMCService) origin(scheme, host string) (string, string) {
	if ms.publicURL != nil {
		return ms.publicURL.Scheme, ms.publicURL.Host
	}
	return scheme, host
}

// BuildForConsumer builds the instance for a pack whose consumer request was
// already authenticated with token. Its pack.toml link carries the same key:
// "public" for a public pack, otherwise that token.
func (ms *MultiMCService) BuildForConsumer(pack tables.Pack, token, scheme, host string) (Archive, response.ServerError) {
	scheme, host = ms.origin(scheme, host)

	return build(packwiz_svc.InstanceZipFileName(pack.Name, pack.Slug), Instance{
		Name:          pack.Name,
		Description:   pack.Description,
		Author:        ms.packwizSvc.GetPackAuthorName(pack.ID),
		Version:       pack.Version,
		MCVersion:     pack.MCVersion,
		Loader:        pack.Loader,
		LoaderVersion: pack.LoaderVersion,
		PackURL:       ms.packwizSvc.ConsumerPackTomlLink(pack, token, scheme, host),
	})
}

func build(filename string, instance Instance) (Archive, response.ServerError) {
	data, err := BuildZip(instance)
	if errors.Is(err, ErrInvalidPack) {
		return Archive{}, response.New(http.StatusUnprocessableEntity, err.Error())
	}
	if err != nil {
		return Archive{}, response.Wrap(err)
	}
	return Archive{Filename: filename, Data: data}, nil
}
