package compat_svc

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/leocov-dev/packwiz-nxt/core"
	// registers the CurseForge MetaDownloader on core.DefaultRegistry
	_ "github.com/leocov-dev/packwiz-nxt/sources"
)

var (
	// ErrUnresolvable means the jar could not be obtained: no usable hash or
	// URL, a download that needs manual action, a network or HTTP failure. The
	// mod is counted as unscanned; it says nothing about the jar itself.
	ErrUnresolvable = errors.New("jar could not be downloaded")
	// ErrTooLarge means the jar exceeds MaxJarSize.
	ErrTooLarge = errors.New("jar exceeds the size limit")
	// ErrHashMismatch means the downloaded bytes do not match the mod's stored hash.
	ErrHashMismatch = errors.New("jar does not match the stored hash")
)

func unresolvable(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrUnresolvable, fmt.Sprintf(format, args...))
}

// TempJar is a downloaded jar in a temporary file. Close closes and deletes
// the file; jars are never retained.
type TempJar struct {
	*os.File
	// Size is the number of bytes in the file.
	Size int64
}

// Close closes the file and removes it from disk.
func (j *TempJar) Close() error {
	name := j.File.Name()
	closeErr := j.File.Close()
	if err := os.Remove(name); err != nil && !errors.Is(err, os.ErrNotExist) {
		return errors.Join(closeErr, err)
	}
	return closeErr
}

// JarFetcher downloads the jar of a mod into a temporary file and checks it
// against the mod's stored hash. The caller must Close the result.
type JarFetcher interface {
	Fetch(ctx context.Context, mod *core.Mod) (*TempJar, error)
}

// HTTPJarFetcher is the JarFetcher used in production. URL downloads go
// through an SSRF-guarded client with size and time limits; CurseForge
// ("metadata:curseforge") mods are resolved through packwiz-nxt's registered
// MetaDownloader, whose URLs come from the CurseForge API and not from user data.
type HTTPJarFetcher struct {
	client *http.Client
}

// NewHTTPJarFetcher builds the production fetcher.
func NewHTTPJarFetcher() *HTTPJarFetcher {
	return &HTTPJarFetcher{client: newGuardedClient(false)}
}

// newTestJarFetcher builds a fetcher whose client may connect to loopback, for
// tests that use httptest servers.
func newTestJarFetcher() *HTTPJarFetcher {
	return &HTTPJarFetcher{client: newGuardedClient(true)}
}

// Fetch implements JarFetcher.
func (f *HTTPJarFetcher) Fetch(ctx context.Context, mod *core.Mod) (*TempJar, error) {
	dl := mod.Download
	if dl.Hash == "" || dl.HashFormat == "" {
		return nil, unresolvable("mod has no stored hash to verify against")
	}
	hasher, err := core.GetHashImpl(dl.HashFormat)
	if err != nil {
		return nil, unresolvable("%v", err)
	}

	ctx, cancel := context.WithTimeout(ctx, DownloadTimeout)
	defer cancel()

	body, length, err := f.open(ctx, mod)
	if err != nil {
		return nil, err
	}
	defer func() { _ = body.Close() }()

	if length > MaxJarSize {
		return nil, fmt.Errorf("%w: %d bytes", ErrTooLarge, length)
	}

	file, err := os.CreateTemp("", "pww-jar-*.jar")
	if err != nil {
		return nil, fmt.Errorf("create temp file: %w", err)
	}
	jar := &TempJar{File: file}
	ok := false
	defer func() {
		if !ok {
			_ = jar.Close()
		}
	}()

	n, err := copyLimited(io.MultiWriter(file, hasher), body, MaxJarSize)
	if err != nil {
		if errors.Is(err, ErrTooLarge) {
			return nil, err
		}
		return nil, unresolvable("read: %v", err)
	}
	if err := verifyHash(hasher, dl.HashFormat, dl.Hash); err != nil {
		return nil, err
	}

	jar.Size = n
	ok = true
	return jar, nil
}

// open starts the download and returns the body and its declared length (-1
// when unknown).
func (f *HTTPJarFetcher) open(ctx context.Context, mod *core.Mod) (io.ReadCloser, int64, error) {
	switch mod.Download.Mode {
	case "", core.ModeURL:
		return f.openURL(ctx, mod.Download.URL)
	case core.ModeCF:
		return openCurseforge(mod)
	default:
		return nil, 0, unresolvable("unsupported download mode %q", mod.Download.Mode)
	}
}

func (f *HTTPJarFetcher) openURL(ctx context.Context, url string) (io.ReadCloser, int64, error) {
	if url == "" {
		return nil, 0, unresolvable("mod has no download URL")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, 0, unresolvable("invalid URL: %v", err)
	}
	if err := checkScheme(req.URL.Scheme); err != nil {
		return nil, 0, unresolvable("%v", err)
	}
	req.Header.Set("User-Agent", core.UserAgent)
	req.Header.Set("Accept", "application/octet-stream")

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, 0, unresolvable("request failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		_ = resp.Body.Close()
		return nil, 0, unresolvable("HTTP status %d", resp.StatusCode)
	}
	return resp.Body, resp.ContentLength, nil
}

func openCurseforge(mod *core.Mod) (io.ReadCloser, int64, error) {
	downloader, ok := core.GetMetaDownloader("curseforge")
	if !ok {
		return nil, 0, unresolvable("curseforge downloads are not available")
	}
	data, err := downloader.GetFilesMetadata([]*core.Mod{mod})
	if err != nil || len(data) != 1 {
		return nil, 0, unresolvable("curseforge file lookup failed: %v", err)
	}
	if manual, _ := data[0].GetManualDownload(); manual {
		return nil, 0, unresolvable("curseforge file must be downloaded manually")
	}
	body, err := data[0].DownloadFile()
	if err != nil {
		return nil, 0, unresolvable("%v", err)
	}
	return body, -1, nil
}

// copyLimited copies src to dst, failing with ErrTooLarge as soon as more than
// max bytes would be copied. Content-Length can lie or be absent, so the bytes
// actually read are counted.
func copyLimited(dst io.Writer, src io.Reader, max int64) (int64, error) {
	n, err := io.Copy(dst, io.LimitReader(src, max+1))
	if err != nil {
		return n, err
	}
	if n > max {
		return n, fmt.Errorf("%w: more than %d bytes", ErrTooLarge, max)
	}
	return n, nil
}

// verifyHash compares the hash computed over the download with the stored one.
func verifyHash(h core.HashStringer, format, expected string) error {
	if got := h.String(); !strings.EqualFold(got, expected) {
		return fmt.Errorf("%w: %s is %s, expected %s", ErrHashMismatch, format, got, expected)
	}
	return nil
}
