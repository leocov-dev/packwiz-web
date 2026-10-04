package packwiz_svc

import (
	"fmt"
	"net/http"
	"sort"
	"time"

	"packwiz-web/internal/tables"
	"packwiz-web/internal/types"
	"packwiz-web/internal/types/dto"
	"packwiz-web/internal/types/response"
)

const (
	changelistDays   = 10
	changelistMonths = 3
	dateLayout       = "2006-01-02"
)

// changelistEvent is a moment the served content of a pack changed: a snapshot
// was recorded, or a revert moved the pack back to an earlier snapshot.
type changelistEvent struct {
	At         time.Time
	SnapshotID uint
}

// changelistBucket groups the events of one period. End is the index of its
// last event; Base the index of the last event before it, or -1.
type changelistBucket struct {
	Period string
	Start  time.Time
	Until  time.Time // last day covered
	End    int
	Base   int
}

func utcDay(t time.Time) time.Time {
	t = t.UTC()
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

// changelistPeriod places a day in its period: the last changelistDays days
// (today included) by day, the rest of the last changelistMonths calendar
// months (this one included) by month, anything older by year. Until is the
// period's last day, cut short where a finer period takes over.
func changelistPeriod(day, now time.Time) (period string, start, until time.Time) {
	today := utcDay(now)
	dayCut := today.AddDate(0, 0, -(changelistDays - 1))
	monthCut := time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, time.UTC).AddDate(0, -(changelistMonths - 1), 0)

	switch {
	case !day.Before(dayCut):
		return dto.ChangelistDay, day, day
	case !day.Before(monthCut):
		start = time.Date(day.Year(), day.Month(), 1, 0, 0, 0, 0, time.UTC)
		until = start.AddDate(0, 1, -1)
		if !until.Before(dayCut) {
			until = dayCut.AddDate(0, 0, -1)
		}
		return dto.ChangelistMonth, start, until
	default:
		start = time.Date(day.Year(), 1, 1, 0, 0, 0, 0, time.UTC)
		until = time.Date(day.Year(), 12, 31, 0, 0, 0, 0, time.UTC)
		if !until.Before(monthCut) {
			until = monthCut.AddDate(0, 0, -1)
		}
		return dto.ChangelistYear, start, until
	}
}

// bucketChangelist groups events (oldest first) into periods, oldest first.
func bucketChangelist(events []changelistEvent, now time.Time) []changelistBucket {
	var buckets []changelistBucket
	for i, e := range events {
		period, start, until := changelistPeriod(utcDay(e.At), now)
		if n := len(buckets); n > 0 && buckets[n-1].Period == period && buckets[n-1].Start.Equal(start) {
			buckets[n-1].End = i
			continue
		}
		base := -1
		if n := len(buckets); n > 0 {
			base = buckets[n-1].End
		}
		buckets = append(buckets, changelistBucket{Period: period, Start: start, Until: until, End: i, Base: base})
	}
	return buckets
}

// changelistModFields are the mod changes worth showing a player, besides an
// update. Hashes, URLs and update metadata only matter to the installer.
var changelistModFields = map[string]bool{
	"name":            true,
	"side":            true,
	"option.optional": true,
	"option.default":  true,
}

// changelistUpdateFields mean the mod's file changed: an update.
var changelistUpdateFields = map[string]bool{
	"version":      true,
	"fileName":     true,
	"download.url": true,
}

func changelistMod(m dto.SnapshotModRef) dto.ChangelistMod {
	return dto.ChangelistMod{Slug: m.Slug, Name: m.Name, Version: m.Version}
}

// buildChangelistEntry turns the change from -> to into a readable entry.
// initial skips listing every mod as added.
func buildChangelistEntry(from, to historyPayload, initial bool) dto.ChangelistEntry {
	diff := diffHistory(from, to)

	entry := dto.ChangelistEntry{
		Initial:  initial,
		ModCount: len(to.Mods),
		Pack:     []dto.SnapshotFieldChange{},
		Added:    []dto.ChangelistMod{},
		Removed:  []dto.ChangelistMod{},
		Changed:  []dto.ChangelistModChange{},
	}

	for _, c := range diff.Pack {
		if c.Field == "description" {
			// the text itself can be long; say that it changed
			c.From, c.To = "", ""
		}
		entry.Pack = append(entry.Pack, c)
	}
	if initial {
		return entry
	}

	for _, m := range diff.Added {
		entry.Added = append(entry.Added, changelistMod(m))
	}
	for _, m := range diff.Removed {
		entry.Removed = append(entry.Removed, changelistMod(m))
	}

	fromBySlug := make(map[string]historyMod, len(from.Mods))
	for _, m := range from.Mods {
		fromBySlug[m.Slug] = m
	}
	for _, c := range diff.Changed {
		change := dto.ChangelistModChange{
			Slug:        c.Slug,
			Name:        c.Name,
			FromVersion: fromBySlug[c.Slug].Version,
			ToVersion:   c.Version,
			Fields:      []string{},
		}
		for _, f := range c.Changes {
			switch {
			case changelistUpdateFields[f.Field]:
				change.Updated = true
			case changelistModFields[f.Field]:
				change.Fields = append(change.Fields, f.Field)
			}
		}
		if change.Updated || len(change.Fields) > 0 {
			entry.Changed = append(entry.Changed, change)
		}
	}

	return entry
}

// isEmptyEntry reports an entry with nothing to show, e.g. changes that
// cancelled out within the period, or only installer-level changes.
func isEmptyEntry(e dto.ChangelistEntry) bool {
	return !e.Initial && len(e.Pack) == 0 && len(e.Added) == 0 && len(e.Removed) == 0 && len(e.Changed) == 0
}

// changelistEvents lists when the served content changed, oldest first: each
// snapshot's creation, and each revert (the abandoned snapshots of one revert
// share abandoned_at and point at the snapshot it restored).
func (ps *PackwizService) changelistEvents(packId uint) ([]changelistEvent, error) {
	var snaps []tables.PackSnapshot
	if err := ps.db.
		Select("id", "seq", "created_at", "abandoned_at", "abandoned_by_revert_to").
		Where("pack_id = ?", packId).
		Order("seq").
		Find(&snaps).Error; err != nil {
		return nil, fmt.Errorf("load snapshots of pack %d: %w", packId, err)
	}

	events := make([]changelistEvent, 0, len(snaps))
	reverts := map[string]changelistEvent{}
	for _, s := range snaps {
		events = append(events, changelistEvent{At: s.CreatedAt, SnapshotID: s.ID})
		if s.AbandonedAt != nil && s.AbandonedByRevertTo != nil {
			key := fmt.Sprintf("%d@%d", *s.AbandonedByRevertTo, s.AbandonedAt.UnixNano())
			reverts[key] = changelistEvent{At: *s.AbandonedAt, SnapshotID: *s.AbandonedByRevertTo}
		}
	}
	for _, r := range reverts {
		events = append(events, r)
	}
	sort.SliceStable(events, func(i, j int) bool { return events[i].At.Before(events[j].At) })

	return events, nil
}

// loadPayloads decodes the payloads of the given snapshots of a pack.
func (ps *PackwizService) loadPayloads(packId uint, ids []uint) (map[uint]historyPayload, error) {
	var snaps []tables.PackSnapshot
	if err := ps.db.
		Select("id", "schema_version", "payload").
		Where("pack_id = ? AND id IN ?", packId, ids).
		Find(&snaps).Error; err != nil {
		return nil, fmt.Errorf("load snapshot payloads of pack %d: %w", packId, err)
	}
	out := make(map[uint]historyPayload, len(snaps))
	for _, s := range snaps {
		p, err := decodeHistoryPayload(s.SchemaVersion, s.Payload)
		if err != nil {
			return nil, err
		}
		out[s.ID] = p
	}
	return out, nil
}

// changelist builds a pack's changelist from its history, newest first.
func (ps *PackwizService) changelist(packId uint, now time.Time) (dto.PackChangelistResponse, error) {
	events, err := ps.changelistEvents(packId)
	if err != nil {
		return dto.PackChangelistResponse{}, err
	}
	buckets := bucketChangelist(events, now)

	idSet := map[uint]struct{}{}
	for _, b := range buckets {
		idSet[events[b.End].SnapshotID] = struct{}{}
	}
	ids := make([]uint, 0, len(idSet))
	for id := range idSet {
		ids = append(ids, id)
	}
	payloads, err := ps.loadPayloads(packId, ids)
	if err != nil {
		return dto.PackChangelistResponse{}, err
	}

	entries := []dto.ChangelistEntry{}
	for _, b := range buckets {
		to, ok := payloads[events[b.End].SnapshotID]
		if !ok {
			continue
		}
		from, initial := emptyHistoryPayload(), b.Base < 0
		if !initial {
			if from, ok = payloads[events[b.Base].SnapshotID]; !ok {
				continue
			}
		}

		entry := buildChangelistEntry(from, to, initial)
		if isEmptyEntry(entry) {
			continue
		}
		entry.Period = b.Period
		entry.Start = b.Start.Format(dateLayout)
		entry.End = b.Until.Format(dateLayout)
		entries = append(entries, entry)
	}

	// newest first
	for i, j := 0, len(entries)-1; i < j; i, j = i+1, j-1 {
		entries[i], entries[j] = entries[j], entries[i]
	}
	return dto.PackChangelistResponse{Entries: entries}, nil
}

// GetChangelist returns a pack's changelist.
func (ps *PackwizService) GetChangelist(packId uint) (dto.PackChangelistResponse, response.ServerError) {
	if _, err := ps.getPackHead(packId); err != nil {
		return dto.PackChangelistResponse{}, response.New(http.StatusNotFound, fmt.Sprintf("pack '%d' not found", packId))
	}
	out, err := ps.changelist(packId, time.Now())
	if err != nil {
		return dto.PackChangelistResponse{}, response.Wrap(err)
	}
	return out, nil
}

// GetPublicChangelist returns the changelist of a public, published pack, or
// a 404 like GetPublicPack.
func (ps *PackwizService) GetPublicChangelist(slug string) (dto.PackChangelistResponse, response.ServerError) {
	var pack tables.Pack
	if err := ps.db.
		Select("id").
		Where("slug = ? AND is_public = ? AND status = ?", slug, true, types.PackStatusPublished).
		First(&pack).Error; err != nil {
		return dto.PackChangelistResponse{}, response.New(http.StatusNotFound, "pack not found")
	}
	out, err := ps.changelist(pack.ID, time.Now())
	if err != nil {
		return dto.PackChangelistResponse{}, response.Wrap(err)
	}
	return out, nil
}

// GetPendingChanges returns what publishing would release: the pack's current
// content compared with the last published state (its history head). A pack
// that was never published has no head and is reported as initial.
func (ps *PackwizService) GetPendingChanges(packId uint) (dto.ChangelistEntry, response.ServerError) {
	pack, err := ps.getPackHead(packId)
	if err != nil {
		return dto.ChangelistEntry{}, response.New(http.StatusNotFound, fmt.Sprintf("pack '%d' not found", packId))
	}

	var full tables.Pack
	if err := ps.db.Unscoped().First(&full, packId).Error; err != nil {
		return dto.ChangelistEntry{}, response.Wrap(err)
	}
	current, loadErr := loadHistoryPayload(ps.db, full)
	if loadErr != nil {
		return dto.ChangelistEntry{}, response.Wrap(loadErr)
	}

	_, headPayload, headErr := loadHeadSnapshot(ps.db, pack)
	if headErr != nil {
		return dto.ChangelistEntry{}, response.Wrap(headErr)
	}

	entry := buildChangelistEntry(headPayload, current, pack.HeadSnapshotID == nil)
	entry.Period = dto.ChangelistPending
	return entry, nil
}
