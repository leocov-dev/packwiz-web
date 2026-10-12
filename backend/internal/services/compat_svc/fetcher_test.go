package compat_svc

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/leocov-dev/packwiz-nxt/core"
)

func sha256Hex(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

func modFor(url string, body []byte) *core.Mod {
	return &core.Mod{Download: core.ModDownload{URL: url, HashFormat: "sha256", Hash: sha256Hex(body)}}
}

func TestCopyLimited(t *testing.T) {
	tests := []struct {
		name    string
		size    int
		max     int64
		wantN   int64
		wantErr error
	}{
		{"empty", 0, 10, 0, nil},
		{"under", 5, 10, 5, nil},
		{"exactly max", 10, 10, 10, nil},
		{"one over", 11, 10, 11, ErrTooLarge},
		{"far over", 1000, 10, 11, ErrTooLarge},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var dst bytes.Buffer
			n, err := copyLimited(&dst, bytes.NewReader(make([]byte, tt.size)), tt.max)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
			if n != tt.wantN {
				t.Errorf("n = %d, want %d", n, tt.wantN)
			}
			if int64(dst.Len()) > tt.max+1 {
				t.Errorf("wrote %d bytes, more than max+1", dst.Len())
			}
		})
	}
}

func TestVerifyHash(t *testing.T) {
	data := []byte("jar bytes")
	h, _ := core.GetHashImpl("sha256")
	_, _ = h.Write(data)
	want := sha256Hex(data)

	tests := []struct {
		name     string
		expected string
		wantErr  bool
	}{
		{"match", want, false},
		{"match upper case", strings.ToUpper(want), false},
		{"mismatch", strings.Repeat("0", 64), true},
		{"empty", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := verifyHash(h, "sha256", tt.expected)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && !errors.Is(err, ErrHashMismatch) {
				t.Errorf("err = %v, want ErrHashMismatch", err)
			}
		})
	}
}

func TestHTTPJarFetcherFetch(t *testing.T) {
	good := []byte("PK-pretend-jar")
	mux := http.NewServeMux()
	mux.HandleFunc("/good.jar", func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write(good) })
	mux.HandleFunc("/other.jar", func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte("something else")) })
	mux.HandleFunc("/missing.jar", func(w http.ResponseWriter, _ *http.Request) { http.NotFound(w, nil) })
	mux.HandleFunc("/redirect.jar", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/good.jar", http.StatusFound)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	tests := []struct {
		name    string
		mod     *core.Mod
		wantErr error
	}{
		{"ok", modFor(srv.URL+"/good.jar", good), nil},
		{"follows redirect", modFor(srv.URL+"/redirect.jar", good), nil},
		{"hash mismatch", modFor(srv.URL+"/other.jar", good), ErrHashMismatch},
		{"http 404", modFor(srv.URL+"/missing.jar", good), ErrUnresolvable},
		{"no url", modFor("", good), ErrUnresolvable},
		{"non-http scheme", modFor("file:///etc/passwd", good), ErrUnresolvable},
		{"no hash", &core.Mod{Download: core.ModDownload{URL: srv.URL + "/good.jar"}}, ErrUnresolvable},
		{"unknown hash format", &core.Mod{Download: core.ModDownload{URL: srv.URL + "/good.jar", HashFormat: "nope", Hash: "x"}}, ErrUnresolvable},
		{"unknown mode", &core.Mod{Download: core.ModDownload{URL: srv.URL + "/good.jar", HashFormat: "sha256", Hash: sha256Hex(good), Mode: "weird"}}, ErrUnresolvable},
	}
	f := newTestJarFetcher()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jar, err := f.Fetch(context.Background(), tt.mod)
			if !errors.Is(err, tt.wantErr) {
				if jar != nil {
					_ = jar.Close()
				}
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}
			if jar.Size != int64(len(good)) {
				t.Errorf("size = %d, want %d", jar.Size, len(good))
			}
			got, _ := io.ReadAll(io.NewSectionReader(jar.File, 0, jar.Size))
			if !bytes.Equal(got, good) {
				t.Errorf("content = %q", got)
			}
			name := jar.Name()
			if err := jar.Close(); err != nil {
				t.Fatalf("close: %v", err)
			}
			if _, err := os.Stat(name); !errors.Is(err, os.ErrNotExist) {
				t.Errorf("temp file still exists after Close: %v", err)
			}
		})
	}
}

func TestHTTPJarFetcherProductionClientBlocksLoopback(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte("x")) }))
	defer srv.Close()

	_, err := NewHTTPJarFetcher().Fetch(context.Background(), modFor(srv.URL, []byte("x")))
	if !errors.Is(err, ErrUnresolvable) || !strings.Contains(err.Error(), errBlockedAddress.Error()) {
		t.Fatalf("err = %v, want unresolvable blocked address", err)
	}
}

func TestHTTPJarFetcherSizeLimit(t *testing.T) {
	// a body larger than the limit with and without Content-Length
	big := make([]byte, MaxJarSize+1024)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/chunked" {
			w.(http.Flusher).Flush() // forces chunked encoding: no Content-Length
		}
		_, _ = w.Write(big)
	}))
	defer srv.Close()

	f := newTestJarFetcher()
	for _, path := range []string{"/declared", "/chunked"} {
		t.Run(path, func(t *testing.T) {
			jar, err := f.Fetch(context.Background(), modFor(srv.URL+path, big))
			if jar != nil {
				_ = jar.Close()
			}
			if !errors.Is(err, ErrTooLarge) {
				t.Fatalf("err = %v, want ErrTooLarge", err)
			}
		})
	}
}
