package middleware

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"packwiz-web/internal/params"
	"packwiz-web/internal/tables"
)

type fakeRecorder struct {
	accesses  []tables.PackAccess
	downloads []tables.InstanceDownload
}

func (f *fakeRecorder) Record(a tables.PackAccess) { f.accesses = append(f.accesses, a) }
func (f *fakeRecorder) RecordInstanceDownload(d tables.InstanceDownload) {
	f.downloads = append(f.downloads, d)
}

func TestPackwizAuditSplitsTomlAndZip(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := &fakeRecorder{}

	router := gin.New()
	group := router.Group(fmt.Sprintf("packwiz/:%s/:%s", params.Token, params.PackSlug))
	group.Use(PackwizAudit(rec))
	ok := func(c *gin.Context) { c.Status(http.StatusOK) }
	group.GET("pack.toml", ok)
	group.GET("index.toml", ok)
	group.GET(fmt.Sprintf("%s/:%s", params.InstanceZipDir, params.InstanceFile), func(c *gin.Context) {
		c.Status(http.StatusForbidden)
	})

	for _, path := range []string{
		"/packwiz/tok/my-pack/pack.toml",
		"/packwiz/tok/my-pack/index.toml",
		"/packwiz/tok/my-pack/multimc/My%20Pack.zip",
	} {
		router.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, path, nil))
	}

	if len(rec.accesses) != 1 || !rec.accesses[0].Success {
		t.Errorf("pack accesses = %+v, want one successful pack.toml", rec.accesses)
	}
	if len(rec.downloads) != 1 {
		t.Fatalf("downloads = %+v, want one", rec.downloads)
	}
	d := rec.downloads[0]
	if d.PackSlug != "my-pack" || d.StatusCode != http.StatusForbidden || d.Success {
		t.Errorf("download = %+v", d)
	}
}
