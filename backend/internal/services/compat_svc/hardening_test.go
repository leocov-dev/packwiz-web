package compat_svc

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/leocov-dev/packwiz-nxt/compat"
	"github.com/leocov-dev/packwiz-nxt/core"
	"github.com/leocov-dev/packwiz-nxt/jarmeta"

	"packwiz-web/internal/jobs"
	"packwiz-web/internal/tables"
)

func TestSanitizeText(t *testing.T) {
	tests := []struct {
		name string
		in   string
		max  int
		want string
	}{
		{"plain", "hello", 10, "hello"},
		{"cut ascii", "hello world", 5, "hello"},
		{"multibyte kept whole", "héllo", 10, "héllo"},
		{"cut never splits rune (2 byte)", "aé", 2, "a"},
		{"cut never splits rune (3 byte)", "a€€", 5, "a€"},
		{"cut exactly at rune end", "a€", 4, "a€"},
		{"4 byte emoji dropped when it does not fit", "ab😀", 4, "ab"},
		{"nul removed", "a\x00b\x00", 10, "ab"},
		{"invalid utf8 replaced", "a\xffb\xc3", 10, "a�b�"},
		{"invalid run collapses to one", strings.Repeat("\xff", 10), 7, "�"},
		{"empty", "", 5, ""},
		{"zero max", "abc", 0, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitizeText(tt.in, tt.max)
			if got != tt.want {
				t.Errorf("sanitizeText(%q, %d) = %q, want %q", tt.in, tt.max, got, tt.want)
			}
			if !utf8.ValidString(got) || strings.ContainsRune(got, 0) || len(got) > tt.max {
				t.Errorf("result %q is not safe to store", got)
			}
		})
	}
}

func TestStripJSONNul(t *testing.T) {
	tests := []struct{ name, in, want string }{
		{"none", `{"a":"b"}`, `{"a":"b"}`},
		{"nul", `{"a":"x\u0000y"}`, `{"a":"xy"}`},
		{"two", `["\u0000","a\u0000"]`, `["","a"]`},
		{"escaped backslash then text", `{"a":"\\u0000"}`, `{"a":"\\u0000"}`},
		{"escaped backslash then nul", `{"a":"\\\u0000"}`, `{"a":"\\"}`},
		{"other escapes untouched", `{"a":"\n\"é"}`, `{"a":"\n\"é"}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := string(stripJSONNul([]byte(tt.in))); got != tt.want {
				t.Errorf("got %s, want %s", got, tt.want)
			}
		})
	}
}

func TestSanitizeMetas(t *testing.T) {
	dirty := []jarmeta.LoaderMeta{{Loader: jarmeta.Loader("fabric"), ModID: "a\x00b", Name: "n\x00", Mods: []jarmeta.Mod{{ID: "a\x00b"}}}}
	got := sanitizeMetas(dirty)
	if got[0].ModID != "ab" || got[0].Name != "n" || got[0].Mods[0].ID != "ab" {
		t.Errorf("not sanitized: %+v", got[0])
	}
	if dirty[0].ModID != "a\x00b" {
		t.Error("input modified")
	}
	clean := []jarmeta.LoaderMeta{{ModID: "ok"}}
	if out := sanitizeMetas(clean); out[0].ModID != "ok" {
		t.Errorf("clean metas changed: %+v", out)
	}
}

func TestRowFromResultSanitizesError(t *testing.T) {
	err := errors.New("bad \x00 jar \xff " + strings.Repeat("€", 2000))
	row := rowFromResult(JarKey{HashFormat: "sha256", Hash: "h"}, jarResult{ParseErr: err})
	if strings.ContainsRune(row.Error, 0) || !utf8.ValidString(row.Error) || len(row.Error) > maxErrorText {
		t.Errorf("unsafe error text (len %d)", len(row.Error))
	}
}

func TestFindingRowsSanitize(t *testing.T) {
	rows := findingRows(1, []compat.Finding{{
		Slug: "s\x00", Severity: compat.SeverityError, Code: compat.CodeMissingDependency,
		Message: "m\xff\x00" + strings.Repeat("é", 3000),
		Detail:  map[string]string{"k\x00": "v\x00\xff", "long": strings.Repeat("a", 2000)},
	}}, nil)
	r := rows[0]
	if r.Slug != "s" || strings.ContainsRune(r.Message, 0) || !utf8.ValidString(r.Message) || len(r.Message) > maxErrorText {
		t.Errorf("unsafe row: %+v", r)
	}
	if r.Detail["k"] != "v�" || len(r.Detail["long"].(string)) != maxDetailText {
		t.Errorf("unsafe detail: %v", r.Detail)
	}
}

func TestClearMissingModIDs(t *testing.T) {
	id := func(n uint) *uint { return &n }
	rows := []tables.PackCompatFinding{{Slug: "a", ModID: id(1)}, {Slug: "b", ModID: id(2)}, {Slug: "c"}}
	got := clearMissingModIDs(rows, []uint{1})
	if got[0].ModID == nil || *got[0].ModID != 1 || got[1].ModID != nil || got[2].ModID != nil {
		t.Errorf("got %+v", got)
	}
	if rows[1].ModID == nil {
		t.Error("input modified")
	}
	if ids := findingModIDs(append(rows, tables.PackCompatFinding{ModID: id(1)})); len(ids) != 2 {
		t.Errorf("ids = %v, want 2 distinct", ids)
	}
}

func TestCatchupNeeded(t *testing.T) {
	now := time.Date(2026, 10, 11, 12, 0, 0, 0, time.UTC)
	run := func(status string, age time.Duration) *tables.PackCompatRun {
		return &tables.PackCompatRun{Status: status, StartedAt: now.Add(-age)}
	}
	tests := []struct {
		name string
		run  *tables.PackCompatRun
		want bool
	}{
		{"never scanned", nil, true},
		{"fresh queued", run(tables.CompatQueued, time.Minute), false},
		{"orphan queued", run(tables.CompatQueued, orphanQueuedAfter+time.Second), true},
		{"fresh running", run(tables.CompatRunning, 5*time.Minute), false},
		{"running just under stale", run(tables.CompatRunning, runStaleAfter-time.Second), false},
		{"stale running", run(tables.CompatRunning, runStaleAfter+time.Second), true},
		{"old done", run(tables.CompatDone, 100*time.Hour), false},
		{"old failed", run(tables.CompatFailed, 100*time.Hour), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := catchupNeeded(tt.run, now); got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFailureTTL(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		want   time.Duration
		wantOK bool
	}{
		{"nil", nil, 0, false},
		{"canceled", context.Canceled, 0, false},
		{"wrapped deadline", fmt.Errorf("x: %w", context.DeadlineExceeded), 0, false},
		{"unresolvable", unresolvable("HTTP status 404"), fetchFailureTTL, true},
		{"too large", fmt.Errorf("%w: 5", ErrTooLarge), fetchFailureTTL, true},
		{"mismatch", fmt.Errorf("%w: a vs b", ErrHashMismatch), hashMismatchTTL, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ttl, ok := failureTTL(tt.err)
			if ttl != tt.want || ok != tt.wantOK {
				t.Errorf("got %v, %v", ttl, ok)
			}
		})
	}
}

func TestNegativeCache(t *testing.T) {
	now := time.Date(2026, 10, 11, 12, 0, 0, 0, time.UTC)
	c := newNegativeCache()
	key := failureKeyOf(&core.Mod{Download: core.ModDownload{URL: "https://x/a.jar", HashFormat: "sha256", Hash: "h"}})
	other := failureKeyOf(&core.Mod{Download: core.ModDownload{URL: "https://y/a.jar", HashFormat: "sha256", Hash: "h"}})

	if found, _ := c.get(key, now); found {
		t.Fatal("empty cache hit")
	}
	c.put(key, unresolvable("HTTP status 404"), now)
	found, err := c.get(key, now.Add(fetchFailureTTL-time.Second))
	if !found || !errors.Is(err, ErrUnresolvable) {
		t.Fatalf("expected remembered failure, got %v %v", found, err)
	}
	if found, _ := c.get(other, now); found {
		t.Error("another URL for the same hash must not be blocked")
	}
	if found, _ := c.get(key, now.Add(fetchFailureTTL)); found {
		t.Error("entry must expire")
	}

	c.put(key, fmt.Errorf("%w: x", ErrHashMismatch), now)
	if found, _ := c.get(key, now.Add(fetchFailureTTL+time.Minute)); !found {
		t.Error("hash mismatch must outlive the fetch failure window")
	}
	c.forget(key)
	if found, _ := c.get(key, now); found {
		t.Error("forget did not drop the entry")
	}
	c.put(key, context.Canceled, now)
	if found, _ := c.get(key, now); found {
		t.Error("context errors must not be remembered")
	}
}

func TestNegativeCacheBounded(t *testing.T) {
	now := time.Date(2026, 10, 11, 12, 0, 0, 0, time.UTC)
	c := newNegativeCache()
	for i := 0; i < negativeCacheMax*2+5; i++ {
		c.put(failureKey{URL: fmt.Sprint(i)}, unresolvable("x"), now)
	}
	if len(c.entries) > negativeCacheMax {
		t.Errorf("%d entries, cap %d", len(c.entries), negativeCacheMax)
	}
	// expired entries are swept before live ones are dropped
	c = newNegativeCache()
	for i := 0; i < negativeCacheMax; i++ {
		c.put(failureKey{URL: fmt.Sprint(i)}, unresolvable("x"), now)
	}
	later := now.Add(2 * fetchFailureTTL)
	c.put(failureKey{URL: "new"}, unresolvable("x"), later)
	if len(c.entries) != 1 {
		t.Errorf("expired entries not swept: %d", len(c.entries))
	}
}

func TestDownloadBudget(t *testing.T) {
	b := newDownloadBudget()
	if !b.allow() || b.exhausted {
		t.Fatal("fresh budget must allow")
	}
	b.spend(ScanDownloadTimeBudget-time.Second, 10)
	if !b.allow() {
		t.Error("budget left, must allow")
	}
	b.spend(time.Second, 10)
	if b.allow() || !b.exhausted {
		t.Error("time budget used up must refuse and remember")
	}

	b = newDownloadBudget()
	b.spend(time.Second, ScanDownloadByteBudget)
	if b.allow() {
		t.Error("byte budget used up must refuse")
	}
	// the budget plus one download must fit inside the job timeout
	if ScanDownloadTimeBudget+DownloadTimeout >= jobs.CompatScanTimeout {
		t.Error("download budget leaves no room for the last download inside the job timeout")
	}
}

type fakeRequester interface {
	RequestScan(ctx context.Context, packID uint) error
}

func TestTypedNilServiceDoesNotPanic(t *testing.T) {
	var svc *CompatService
	var req fakeRequester = svc // non-nil interface holding a nil pointer
	if err := req.RequestScan(context.Background(), 1); !errors.Is(err, jobs.ErrClientUnavailable) {
		t.Errorf("RequestScan err = %v", err)
	}
	if _, err := svc.RequestManualScan(context.Background(), 1); !errors.Is(err, jobs.ErrClientUnavailable) {
		t.Errorf("RequestManualScan err = %v", err)
	}
	if _, err := svc.EnqueueMissing(context.Background()); !errors.Is(err, jobs.ErrClientUnavailable) {
		t.Errorf("EnqueueMissing err = %v", err)
	}
	if _, err := svc.GetCompat(context.Background(), 1); !errors.Is(err, jobs.ErrClientUnavailable) {
		t.Errorf("GetCompat err = %v", err)
	}
	if err := svc.RunScan(context.Background(), jobs.CompatScanArgs{PackID: 1}, 1); !errors.Is(err, jobs.ErrClientUnavailable) {
		t.Errorf("RunScan err = %v", err)
	}
	if _, err := svc.PruneJarMeta(context.Background()); !errors.Is(err, jobs.ErrClientUnavailable) {
		t.Errorf("PruneJarMeta err = %v", err)
	}
	if err := svc.PurgeTempFiles(); !errors.Is(err, jobs.ErrClientUnavailable) {
		t.Errorf("PurgeTempFiles err = %v", err)
	}
}
