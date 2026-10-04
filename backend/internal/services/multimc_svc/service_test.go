package multimc_svc

import (
	"net/http"
	"testing"

	"packwiz-web/internal/types/response"
)

func TestOriginPrefersPublicURL(t *testing.T) {
	ms := NewMultiMCService(nil, nil, "https://packwiz.example.com")
	scheme, host := ms.origin("http", "192.168.1.10:8080")
	if scheme != "https" || host != "packwiz.example.com" {
		t.Errorf("got %s://%s", scheme, host)
	}
}

func TestOriginFallsBackToRequest(t *testing.T) {
	for _, raw := range []string{"", "not a url", "packwiz.example.com"} {
		ms := NewMultiMCService(nil, nil, raw)
		scheme, host := ms.origin("http", "localhost:8080")
		if scheme != "http" || host != "localhost:8080" {
			t.Errorf("%q: got %s://%s", raw, scheme, host)
		}
	}
}

func TestBuildMapsInvalidPackTo422(t *testing.T) {
	i := fabricInstance()
	i.MCVersion = ""
	_, err := build("blarg.zip", i)
	if err == nil {
		t.Fatal("expected error")
	}
	httpErr, ok := err.(*response.HttpError)
	if !ok || httpErr.Code != http.StatusUnprocessableEntity {
		t.Errorf("err = %v, want a 422", err)
	}
}

func TestBuildUsesFilename(t *testing.T) {
	archive, err := build("blarg.zip", fabricInstance())
	if err != nil {
		t.Fatal(err)
	}
	if archive.Filename != "blarg.zip" || len(archive.Data) == 0 {
		t.Errorf("archive = %s, %d bytes", archive.Filename, len(archive.Data))
	}
}
