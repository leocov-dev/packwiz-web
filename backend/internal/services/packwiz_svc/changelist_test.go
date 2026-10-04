package packwiz_svc

import (
	"testing"
	"time"

	"packwiz-web/internal/types/dto"
)

func day(s string) time.Time {
	t, err := time.Parse(dateLayout, s)
	if err != nil {
		panic(err)
	}
	return t
}

func TestChangelistPeriod(t *testing.T) {
	now := day("2026-10-04").Add(15 * time.Hour)
	for _, tc := range []struct {
		day, period, start, until string
	}{
		{"2026-10-04", dto.ChangelistDay, "2026-10-04", "2026-10-04"},
		{"2026-09-25", dto.ChangelistDay, "2026-09-25", "2026-09-25"},   // 10th day back
		{"2026-09-24", dto.ChangelistMonth, "2026-09-01", "2026-09-24"}, // cut by the day range
		{"2026-08-31", dto.ChangelistMonth, "2026-08-01", "2026-08-31"},
		{"2026-07-31", dto.ChangelistYear, "2026-01-01", "2026-07-31"}, // cut by the month range
		{"2025-03-02", dto.ChangelistYear, "2025-01-01", "2025-12-31"},
	} {
		period, start, until := changelistPeriod(day(tc.day), now)
		if period != tc.period || start.Format(dateLayout) != tc.start || until.Format(dateLayout) != tc.until {
			t.Errorf("%s: got %s %s..%s, want %s %s..%s", tc.day, period, start.Format(dateLayout), until.Format(dateLayout), tc.period, tc.start, tc.until)
		}
	}
}

func TestBucketChangelist(t *testing.T) {
	now := day("2026-10-04")
	events := []changelistEvent{
		{At: day("2025-02-01"), SnapshotID: 1},
		{At: day("2025-06-01"), SnapshotID: 2},
		{At: day("2026-09-02"), SnapshotID: 3},
		{At: day("2026-10-03").Add(time.Hour), SnapshotID: 4},
		{At: day("2026-10-03").Add(2 * time.Hour), SnapshotID: 5},
	}
	got := bucketChangelist(events, now)

	want := []struct {
		period    string
		start     string
		end, base int
	}{
		{dto.ChangelistYear, "2025-01-01", 1, -1},
		{dto.ChangelistMonth, "2026-09-01", 2, 1},
		{dto.ChangelistDay, "2026-10-03", 4, 2},
	}
	if len(got) != len(want) {
		t.Fatalf("got %d buckets, want %d: %+v", len(got), len(want), got)
	}
	for i, w := range want {
		b := got[i]
		if b.Period != w.period || b.Start.Format(dateLayout) != w.start || b.End != w.end || b.Base != w.base {
			t.Errorf("bucket %d = %+v, want %+v", i, b, w)
		}
	}
}

func payloadWith(mc, loader string, mods ...historyMod) historyPayload {
	p := emptyHistoryPayload()
	p.Pack.MCVersion, p.Pack.Loader = mc, loader
	p.Mods = mods
	return p
}

func TestBuildChangelistEntry(t *testing.T) {
	sodium1 := historyMod{Slug: "sodium", Name: "Sodium", Version: "0.5", FileName: "sodium-0.5.jar", Side: "client"}
	sodium2 := sodium1
	sodium2.Version, sodium2.FileName = "0.6", "sodium-0.6.jar"
	lithium := historyMod{Slug: "lithium", Name: "Lithium", Version: "1.0"}
	rehashed := lithium
	rehashed.Download.Hash = "new-hash" // installer-level only
	jei := historyMod{Slug: "jei", Name: "JEI", Version: "2", Side: "both"}
	jeiClient := jei
	jeiClient.Side = "client"
	iris := historyMod{Slug: "iris", Name: "Iris", Version: "1.7"}

	from := payloadWith("1.21.1", "fabric", sodium1, lithium, jei)
	from.Pack.Description = "old"
	to := payloadWith("26.1", "fabric", sodium2, rehashed, jeiClient, iris)
	to.Pack.Description = "new"

	e := buildChangelistEntry(from, to, false)

	if e.ModCount != 4 || e.Initial {
		t.Errorf("modCount=%d initial=%v", e.ModCount, e.Initial)
	}
	fields := map[string]dto.SnapshotFieldChange{}
	for _, c := range e.Pack {
		fields[c.Field] = c
	}
	if c := fields["mcVersion"]; c.From != "1.21.1" || c.To != "26.1" {
		t.Errorf("mcVersion change = %+v", c)
	}
	if c, ok := fields["description"]; !ok || c.From != "" || c.To != "" {
		t.Errorf("description change should be present without text: %+v", c)
	}
	if len(e.Added) != 1 || e.Added[0].Slug != "iris" || len(e.Removed) != 0 {
		t.Errorf("added=%+v removed=%+v", e.Added, e.Removed)
	}

	changed := map[string]dto.ChangelistModChange{}
	for _, c := range e.Changed {
		changed[c.Slug] = c
	}
	if c := changed["sodium"]; !c.Updated || c.FromVersion != "0.5" || c.ToVersion != "0.6" {
		t.Errorf("sodium = %+v", c)
	}
	if c := changed["jei"]; c.Updated || len(c.Fields) != 1 || c.Fields[0] != "side" {
		t.Errorf("jei = %+v", c)
	}
	if _, ok := changed["lithium"]; ok {
		t.Error("a hash-only change must not be listed")
	}
}

func TestBuildChangelistEntryInitial(t *testing.T) {
	to := payloadWith("26.1", "neoforge", historyMod{Slug: "a"}, historyMod{Slug: "b"})
	e := buildChangelistEntry(emptyHistoryPayload(), to, true)
	if !e.Initial || e.ModCount != 2 || len(e.Added) != 0 {
		t.Errorf("initial entry = %+v", e)
	}
	if isEmptyEntry(e) {
		t.Error("an initial entry is never empty")
	}
}

func TestIsEmptyEntry(t *testing.T) {
	p := payloadWith("26.1", "fabric", historyMod{Slug: "a", Version: "1"})
	if !isEmptyEntry(buildChangelistEntry(p, p, false)) {
		t.Error("identical states should give an empty entry")
	}
}
