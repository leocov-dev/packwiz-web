package compat_svc

import (
	"context"
	"errors"
	"fmt"
	"time"

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
		log.Error(fmt.Errorf("compat scan failed for pack %d: %w", args.PackID, err))
		s.markFailed(args.PackID, err)
	}
	return nil
}

// markFailed records a failed run. Best effort: errors are only logged. It
// uses a fresh context because the job's may be cancelled (timeout, shutdown).
func (s *CompatService) markFailed(packID uint, cause error) {
	ctx, cancel := context.WithTimeout(context.Background(), failureWriteTimeout)
	defer cancel()

	if err := s.db.WithContext(ctx).Model(&tables.PackCompatRun{}).Where("pack_id = ?", packID).
		Updates(map[string]any{
			"status":      tables.CompatFailed,
			"finished_at": s.now(),
			"error":       truncate(cause.Error(), maxErrorText),
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
	unscanned := 0
	for _, m := range jars {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("scan interrupted: %w", err)
		}
		modIDBySlug[m.Slug] = m.ID

		res, err := s.resolveJar(ctx, m)
		switch {
		case err == nil:
			entries = append(entries, entryFor(m, res))
		case errors.Is(err, ErrHashMismatch):
			// the jar at the URL is not the one the pack recorded: report it
			// as an unreadable jar instead of trusting its contents
			entries = append(entries, entryFor(m, jarResult{ParseErr: err}))
		case ctx.Err() != nil:
			return fmt.Errorf("scan interrupted: %w", ctx.Err())
		default:
			// download or lookup problems say nothing about the jar: count the
			// mod as unscanned and emit no finding for it
			log.Debug(fmt.Errorf("compat scan: mod %s of pack %d not scanned: %w", m.Slug, packID, err))
			unscanned++
		}
	}

	findings := compat.Check(compat.Target{
		MinecraftVersion: pack.MCVersion,
		Loader:           jarmeta.Loader(pack.Loader),
		LoaderVersion:    pack.LoaderVersion,
	}, entries)
	findings = softenForUnscanned(findings, unscanned)
	counts := countFindings(findings)
	rows := findingRows(packID, findings, modIDBySlug)

	return s.store(ctx, packID, targetHash, counts, unscanned, rows)
}

// store replaces the pack's findings and completes its run in one transaction.
func (s *CompatService) store(ctx context.Context, packID uint, targetHash string, counts FindingCounts, unscanned int, rows []tables.PackCompatFinding) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("pack_id = ?", packID).Delete(&tables.PackCompatFinding{}).Error; err != nil {
			return err
		}
		if len(rows) > 0 {
			if err := tx.CreateInBatches(&rows, 200).Error; err != nil {
				return err
			}
		}
		// a status flipped back to queued while this scan ran (see enqueue)
		// asks for a rescan: keep it
		return tx.Model(&tables.PackCompatRun{}).Where("pack_id = ?", packID).
			Updates(map[string]any{
				"status":          gorm.Expr("CASE WHEN status = ? THEN status ELSE ? END", tables.CompatQueued, tables.CompatDone),
				"target_hash":     targetHash,
				"error_count":     counts.Errors,
				"warning_count":   counts.Warnings,
				"info_count":      counts.Infos,
				"unscanned_count": unscanned,
				"finished_at":     s.now(),
				"error":           "",
			}).Error
	})
}

// resolveJar returns the parsed metadata of a mod's jar: from the mod_jar_meta
// cache when a usable row exists, else by downloading, parsing and caching it.
func (s *CompatService) resolveJar(ctx context.Context, m tables.Mod) (jarResult, error) {
	key := jarKeyOf(m)
	if key.Hash == "" || key.HashFormat == "" {
		return jarResult{}, fmt.Errorf("%w: mod has no stored hash", ErrUnresolvable)
	}

	if row, ok, err := s.cachedRow(ctx, key); err != nil {
		return jarResult{}, err
	} else if ok {
		s.touch(ctx, key)
		return resultFromRow(row), nil
	}

	// two packs sharing a jar download it once: later callers wait for the
	// first and take its result
	v, err, _ := s.flight.Do(key.HashFormat+":"+key.Hash, func() (any, error) {
		return s.downloadAndCache(ctx, m, key)
	})
	if err != nil {
		return jarResult{}, err
	}
	return v.(jarResult), nil
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

// touch marks a cache row as used (GORM does not maintain last_used_at on
// update). Best effort: a missed touch only makes the row prunable sooner.
func (s *CompatService) touch(ctx context.Context, key JarKey) {
	if err := s.db.WithContext(ctx).Model(&tables.ModJarMeta{}).
		Where("hash_format = ? AND hash = ?", key.HashFormat, key.Hash).
		Update("last_used_at", s.now()).Error; err != nil {
		log.Error("failed to touch jar metadata:", err)
	}
}

// downloadAndCache fetches, parses and caches one jar. The temporary file is
// deleted before returning. Parse errors are cached too (with whatever
// metadata did parse); fetch errors are not.
func (s *CompatService) downloadAndCache(ctx context.Context, m tables.Mod, key JarKey) (jarResult, error) {
	jar, err := s.fetcher.Fetch(ctx, m.AsMeta())
	if err != nil {
		return jarResult{}, err
	}
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
