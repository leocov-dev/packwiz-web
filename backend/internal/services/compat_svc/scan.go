package compat_svc

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/leocov-dev/packwiz-nxt/compat"
	"github.com/leocov-dev/packwiz-nxt/jarmeta"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"packwiz-web/internal/jobs"
	"packwiz-web/internal/log"
	"packwiz-web/internal/tables"
)

// failureWriteTimeout bounds recording a failed run after the job context is gone.
const failureWriteTimeout = 10 * time.Second

// RunScan implements jobs.CompatScanner. Failures are recorded on the run row
// and not returned, so River does not retry and download everything again.
//
// The scan does not lock the pack. The target hash stored with the result is
// computed from the data the scan read, so if the pack changed meanwhile the
// badge reads stale (and the change's own trigger queues a follow-up scan).
func (s *CompatService) RunScan(ctx context.Context, args jobs.CompatScanArgs, jobID int64) error {
	if s == nil {
		return jobs.ErrClientUnavailable
	}
	running := tables.PackCompatRun{PackID: args.PackID, JobID: jobID, Status: tables.CompatRunning, StartedAt: s.now()}
	if err := s.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "pack_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"job_id", "status", "started_at"}),
	}).Create(&running).Error; err != nil {
		// e.g. the pack was deleted after enqueue (FK error)
		log.Error(fmt.Errorf("mark compat scan of pack %d running: %w", args.PackID, err))
		return nil
	}

	if err := s.scan(ctx, args.PackID); err != nil {
		// the raw cause stays in the log: it can name tables and constraints
		log.Error(fmt.Errorf("compat scan failed for pack %d: %w", args.PackID, err))
		s.markFailed(args.PackID, failureReason(err))
	}
	return nil
}

// PurgeTempFiles deletes jar downloads orphaned by a killed process. Call it
// once at worker start, before scans run. It is a no-op for fetchers that keep
// no files.
func (s *CompatService) PurgeTempFiles() error {
	if s == nil {
		return jobs.ErrClientUnavailable
	}
	if p, ok := s.fetcher.(interface{ PurgeTempDir() error }); ok {
		return p.PurgeTempDir()
	}
	return nil
}

// isContextErr reports whether err is a cancelled or expired context.
func isContextErr(err error) bool {
	return errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)
}

// Failure reasons stored with a failed run and shown to every viewer of the
// pack. They are fixed texts: the raw cause can carry database or file system
// details and only goes to the log.
const (
	failureTimedOut = "scan timed out"
	failureFailed   = "scan failed"
)

// failureReason classifies a scan error as a user-safe reason.
func failureReason(cause error) string {
	if isContextErr(cause) {
		return failureTimedOut
	}
	return failureFailed
}

// markFailed records a failed run with a user-safe reason. Best effort: errors are only logged. It
// uses a fresh context because the job's may be cancelled (timeout, shutdown).
// A run flipped back to queued while the scan ran (see enqueue) stays queued:
// it still needs its rescan, and the catch-up job will queue it.
func (s *CompatService) markFailed(packID uint, reason string) {
	ctx, cancel := context.WithTimeout(context.Background(), failureWriteTimeout)
	defer cancel()

	if err := s.db.WithContext(ctx).Model(&tables.PackCompatRun{}).
		Where("pack_id = ? AND status <> ?", packID, tables.CompatQueued).
		Updates(map[string]any{
			"status":      tables.CompatFailed,
			"finished_at": s.now(),
			"error":       reason,
		}).Error; err != nil {
		log.Error("failed to record failed compat scan:", err)
	}
}

func (s *CompatService) scan(ctx context.Context, packID uint) error {
	db := s.db.WithContext(ctx)

	pack, mods, err := loadScanInput(db, packID, false)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("pack is archived or deleted")
		}
		return err
	}
	jars := jarMods(mods)
	targetHash := TargetHash(targetInputOf(pack, mods))

	entries := make([]compat.Entry, 0, len(jars))
	modIDBySlug := make(map[string]uint, len(jars))
	var skipped []compat.Finding
	budget := newDownloadBudget()
	used := make(map[JarKey]struct{}, len(jars))
	for _, m := range jars {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("scan interrupted: %w", err)
		}
		modIDBySlug[m.Slug] = m.ID

		res, err := s.resolveJar(ctx, m, budget)
		switch {
		case err == nil:
			used[jarKeyOf(m)] = struct{}{}
			entries = append(entries, entryFor(m, res))
		case errors.Is(err, ErrHashMismatch):
			// the jar at the URL is not the one the pack recorded: report it
			// as an unreadable jar instead of trusting its contents
			entries = append(entries, entryFor(m, jarResult{ParseErr: err}))
		case ctx.Err() != nil:
			return fmt.Errorf("scan interrupted: %w", ctx.Err())
		case errors.Is(err, errBudgetExhausted):
			skipped = append(skipped, notScanned(m.Slug, "download budget exhausted; the next scan continues"))
		default:
			// download or lookup problems say nothing about the jar: count the
			// mod as unscanned and emit no finding for it
			log.Debug(fmt.Errorf("compat scan: mod %s of pack %d not scanned: %w", m.Slug, packID, err))
			skipped = append(skipped, notScanned(m.Slug, "the jar could not be downloaded"))
		}
	}
	unscanned := len(skipped)

	s.touch(ctx, used)

	findings := compat.Check(compat.Target{
		MinecraftVersion: pack.MCVersion,
		Loader:           jarmeta.Loader(pack.Loader),
		LoaderVersion:    pack.LoaderVersion,
	}, entries)
	findings = softenForUnscanned(findings, unscanned)
	findings = append(findings, skipped...)
	counts := countFindings(findings)
	rows := findingRows(packID, findings, modIDBySlug)

	// Out of download budget: the result covers what was scanned, and the run
	// stays queued so the catch-up job continues it (downloaded jars are cached
	// by then; failed URLs are remembered, so the next pass is shorter).
	return s.store(ctx, packID, targetHash, counts, unscanned, rows, budget.exhausted)
}

// store replaces the pack's findings and completes its run in one transaction.
// With partial the run is left queued (with a fresh started_at, so catch-up
// waits orphanQueuedAfter) instead of done.
//
// A mod deleted while the scan ran would make its findings violate the mod_id
// foreign key. Those findings keep a NULL mod_id (see clearMissingModIDs); if
// a mod vanishes between that check and the insert, the transaction is retried
// once.
func (s *CompatService) store(ctx context.Context, packID uint, targetHash string, counts FindingCounts, unscanned int, rows []tables.PackCompatFinding, partial bool) error {
	err := s.storeOnce(ctx, packID, targetHash, counts, unscanned, rows, partial)
	if isForeignKeyViolation(err) {
		err = s.storeOnce(ctx, packID, targetHash, counts, unscanned, rows, partial)
	}
	return err
}

func (s *CompatService) storeOnce(ctx context.Context, packID uint, targetHash string, counts FindingCounts, unscanned int, rows []tables.PackCompatFinding, partial bool) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("pack_id = ?", packID).Delete(&tables.PackCompatFinding{}).Error; err != nil {
			return err
		}
		if len(rows) > 0 {
			var existing []uint
			if ids := findingModIDs(rows); len(ids) > 0 {
				if err := tx.Model(&tables.Mod{}).Where("id IN ?", ids).Pluck("id", &existing).Error; err != nil {
					return err
				}
			}
			rows = clearMissingModIDs(rows, existing)
			if err := tx.CreateInBatches(&rows, 200).Error; err != nil {
				return err
			}
		}
		final := tables.CompatDone
		now := s.now()
		update := map[string]any{
			"target_hash":     targetHash,
			"error_count":     counts.Errors,
			"warning_count":   counts.Warnings,
			"info_count":      counts.Infos,
			"unscanned_count": unscanned,
			"finished_at":     now,
			"error":           "",
		}
		if partial {
			final = tables.CompatQueued
			update["started_at"] = now
		}
		// a status flipped back to queued while this scan ran (see enqueue)
		// asks for a rescan: keep it
		update["status"] = gorm.Expr("CASE WHEN status = ? THEN status ELSE ? END", tables.CompatQueued, final)
		return tx.Model(&tables.PackCompatRun{}).Where("pack_id = ?", packID).Updates(update).Error
	})
}

// pgForeignKeyViolation is the Postgres error code for a foreign key violation.
const pgForeignKeyViolation = "23503"

// isForeignKeyViolation reports whether err is a Postgres foreign key violation.
func isForeignKeyViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == pgForeignKeyViolation
}

// findingModIDs lists the distinct mod ids the rows refer to.
func findingModIDs(rows []tables.PackCompatFinding) []uint {
	seen := make(map[uint]struct{}, len(rows))
	var ids []uint
	for _, r := range rows {
		if r.ModID == nil {
			continue
		}
		if _, ok := seen[*r.ModID]; !ok {
			seen[*r.ModID] = struct{}{}
			ids = append(ids, *r.ModID)
		}
	}
	return ids
}

// clearMissingModIDs returns rows with a NULL mod_id wherever the mod is not in
// existing. The input rows are not modified.
func clearMissingModIDs(rows []tables.PackCompatFinding, existing []uint) []tables.PackCompatFinding {
	have := make(map[uint]struct{}, len(existing))
	for _, id := range existing {
		have[id] = struct{}{}
	}
	out := make([]tables.PackCompatFinding, len(rows))
	copy(out, rows)
	for i := range out {
		if out[i].ModID == nil {
			continue
		}
		if _, ok := have[*out[i].ModID]; !ok {
			out[i].ModID = nil
		}
	}
	return out
}

// resolveJar returns the parsed metadata of a mod's jar: from the mod_jar_meta
// cache when a usable row exists, else by downloading, parsing and caching it
// (within the scan's download budget).
func (s *CompatService) resolveJar(ctx context.Context, m tables.Mod, budget *downloadBudget) (jarResult, error) {
	key := jarKeyOf(m)
	if key.Hash == "" || key.HashFormat == "" {
		return jarResult{}, fmt.Errorf("%w: mod has no stored hash", ErrUnresolvable)
	}

	if row, ok, err := s.cachedRow(ctx, key); err != nil {
		return jarResult{}, err
	} else if ok {
		return resultFromRow(row), nil
	}

	// Callers resolving the same jar at once share one download. The flight
	// runs with its first caller's context and budget, so a follower can get
	// that caller's cancellation as the result; it then retries once under
	// its own context.
	for attempt := 0; ; attempt++ {
		v, err, _ := s.flight.Do(key.HashFormat+":"+key.Hash, func() (any, error) {
			// a flight that finished just before this one started has cached it
			if row, ok, err := s.cachedRow(ctx, key); err != nil {
				return jarResult{}, err
			} else if ok {
				return resultFromRow(row), nil
			}
			return s.downloadAndCache(ctx, m, key, budget)
		})
		if err != nil {
			if attempt == 0 && ctx.Err() == nil && isContextErr(err) {
				continue
			}
			return jarResult{}, err
		}
		return v.(jarResult), nil
	}
}

// cachedRow loads the usable cache row of a jar, if any.
func (s *CompatService) cachedRow(ctx context.Context, key JarKey) (tables.ModJarMeta, bool, error) {
	var rows []tables.ModJarMeta
	if err := s.db.WithContext(ctx).
		Where("hash_format = ? AND hash = ? AND parser_version = ?", key.HashFormat, key.Hash, ParserVersion).
		Limit(1).Find(&rows).Error; err != nil {
		return tables.ModJarMeta{}, false, fmt.Errorf("look up jar metadata: %w", err)
	}
	if len(rows) == 0 || !cacheUsable(rows[0], s.now()) {
		return tables.ModJarMeta{}, false, nil
	}
	return rows[0], true, nil
}

// touch marks the cache rows a scan used as used, in one statement (GORM does
// not maintain last_used_at on update). Best effort: a missed touch only
// makes the rows prunable sooner.
func (s *CompatService) touch(ctx context.Context, keys map[JarKey]struct{}) {
	if len(keys) == 0 {
		return
	}
	tuples := make([][]any, 0, len(keys))
	for k := range keys {
		tuples = append(tuples, []any{k.HashFormat, k.Hash})
	}
	for start := 0; start < len(tuples); start += touchBatch {
		batch := tuples[start:min(start+touchBatch, len(tuples))]
		if err := s.db.WithContext(ctx).Exec(
			`UPDATE mod_jar_meta SET last_used_at = ? WHERE (hash_format, hash) IN ?`, s.now(), batch).Error; err != nil {
			log.Error("failed to touch jar metadata:", err)
			return
		}
	}
}

// touchBatch is how many cache rows one touch UPDATE covers.
const touchBatch = 500

// downloadAndCache fetches, parses and caches one jar. The temporary file is
// deleted before returning. Parse errors are cached too (with whatever
// metadata did parse). Fetch errors are remembered in memory for a while
// (negativeCache), not in mod_jar_meta, and a download is only started while
// the scan has budget. The budget is checked before a download, so the last
// one may overrun it by up to DownloadTimeout; ScanDownloadTimeBudget leaves
// room for that within the job timeout.
func (s *CompatService) downloadAndCache(ctx context.Context, m tables.Mod, key JarKey, budget *downloadBudget) (jarResult, error) {
	mod := m.AsMeta()
	fk := failureKeyOf(mod)
	if found, err := s.failures.get(fk, s.now()); found {
		return jarResult{}, err
	}
	if !budget.allow() {
		return jarResult{}, errBudgetExhausted
	}

	started := s.now()
	jar, err := s.fetcher.Fetch(ctx, mod)
	if err != nil {
		budget.spend(s.now().Sub(started), 0)
		if ctx.Err() == nil { // a cancelled scan says nothing about the URL
			s.failures.put(fk, err, s.now())
		}
		return jarResult{}, err
	}
	budget.spend(s.now().Sub(started), jar.Size)
	s.failures.forget(fk)

	// remove the jar as soon as it is parsed
	metas, parseErr := jarmeta.Parse(jar.File, jar.Size)
	if err := jar.Close(); err != nil {
		log.Error("failed to remove temporary jar:", err)
	}

	res := jarResult{Metas: metas, ParseErr: parseErr}
	row := rowFromResult(key, res)
	if err := s.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "hash_format"}, {Name: "hash"}},
		UpdateAll: true,
	}).Create(&row).Error; err != nil {
		return jarResult{}, fmt.Errorf("cache jar metadata: %w", err)
	}
	return res, nil
}
