package pack_access_svc

import (
	"context"
	"fmt"
	"sync"
	"time"

	"gorm.io/gorm"

	"packwiz-web/internal/log"
	"packwiz-web/internal/tables"
	"packwiz-web/internal/types/dto"
	"packwiz-web/internal/types/response"
)

const topLimit = 10

// recordBuffer bounds how many access records can wait for a DB write.
const recordBuffer = 1024

// Source is a consumer access table the system-wide metrics can read. Its value
// is spliced into SQL as a table name, so only these constants may be used.
type Source string

const (
	// SourcePackToml is pack.toml fetches: pack syncs.
	SourcePackToml Source = "pack_accesses"
	// SourceInstanceZip is MultiMC / Prism instance zip downloads.
	SourceInstanceZip Source = "instance_downloads"
)

type PackAccessService struct {
	db *gorm.DB
	// records holds *tables.PackAccess and *tables.InstanceDownload values.
	records chan any
	start   sync.Once
}

// NewPackAccessService builds the service. The background writer starts on the
// first Record call and lives for the life of the process.
func NewPackAccessService(db *gorm.DB) *PackAccessService {
	return &PackAccessService{db: db, records: make(chan any, recordBuffer)}
}

// Record queues a pack.toml access record for writing. It never blocks the
// request: when the queue is full the record is dropped and logged.
func (s *PackAccessService) Record(access tables.PackAccess) {
	s.enqueue(&access, access.PackSlug)
}

// RecordInstanceDownload queues an instance zip download record, like Record.
func (s *PackAccessService) RecordInstanceDownload(download tables.InstanceDownload) {
	s.enqueue(&download, download.PackSlug)
}

func (s *PackAccessService) enqueue(record any, slug string) {
	s.start.Do(func() { go s.write() })

	select {
	case s.records <- record:
	default:
		log.Warn("pack access queue full, dropping record for", slug)
	}
}

func (s *PackAccessService) write() {
	for record := range s.records {
		if err := s.db.Create(record).Error; err != nil {
			log.Error(fmt.Sprintf("Failed to create pack access record: %s", err))
		}
	}
}

// PruneOlderThan deletes pack access and instance download rows created before
// cutoff and returns how many were removed.
func (s *PackAccessService) PruneOlderThan(ctx context.Context, cutoff time.Time) (int64, error) {
	var total int64
	for _, model := range []any{&tables.PackAccess{}, &tables.InstanceDownload{}} {
		result := s.db.WithContext(ctx).Where("created_at < ?", cutoff).Delete(model)
		if result.Error != nil {
			return total, result.Error
		}
		total += result.RowsAffected
	}
	return total, nil
}

// PackSeries returns daily successful-access counts for a pack.
func (s *PackAccessService) PackSeries(packId uint, days int) ([]dto.AccessDay, response.ServerError) {
	now := time.Now()
	var sparse []dto.AccessDay
	err := s.db.Raw(`
		SELECT to_char(created_at AT TIME ZONE 'UTC', 'YYYY-MM-DD') AS date,
		       COUNT(*) AS success
		FROM pack_accesses
		WHERE pack_id = ? AND success AND created_at >= ?
		GROUP BY 1`, packId, rangeStart(now, days)).Scan(&sparse).Error
	if err != nil {
		return nil, response.Wrap(fmt.Errorf("pack access series: %w", err))
	}
	return FillDays(sparse, now, days), nil
}

// PackSummary returns successful-access metrics for a pack.
func (s *PackAccessService) PackSummary(packId uint, days int) (*dto.PackAccessSummary, response.ServerError) {
	series, serr := s.PackSeries(packId, days)
	if serr != nil {
		return nil, serr
	}

	since := rangeStart(time.Now(), days)
	summary := &dto.PackAccessSummary{Days: days, Series: series}

	err := s.db.Raw(`
		SELECT COUNT(*) AS total,
		       COUNT(DISTINCT ip_address) AS unique_ips,
		       COUNT(DISTINCT user_id) AS unique_users
		FROM pack_accesses
		WHERE pack_id = ? AND success AND created_at >= ?`, packId, since).
		Row().Scan(&summary.Total, &summary.UniqueIps, &summary.UniqueUsers)
	if err != nil {
		return nil, response.Wrap(fmt.Errorf("pack access totals: %w", err))
	}

	summary.TopUsers = []dto.AccessUserCount{}
	err = s.db.Raw(`
		SELECT u.id AS user_id, u.username, COUNT(*) AS count
		FROM pack_accesses a JOIN users u ON u.id = a.user_id
		WHERE a.pack_id = ? AND a.success AND a.created_at >= ?
		GROUP BY u.id, u.username
		ORDER BY count DESC, u.username LIMIT ?`, packId, since, topLimit).Scan(&summary.TopUsers).Error
	if err != nil {
		return nil, response.Wrap(fmt.Errorf("pack access top users: %w", err))
	}

	summary.TopIps = []dto.AccessIpCount{}
	err = s.db.Raw(`
		SELECT ip_address, COUNT(*) AS count
		FROM pack_accesses
		WHERE pack_id = ? AND success AND created_at >= ?
		GROUP BY ip_address
		ORDER BY count DESC, ip_address LIMIT ?`, packId, since, topLimit).Scan(&summary.TopIps).Error
	if err != nil {
		return nil, response.Wrap(fmt.Errorf("pack access top ips: %w", err))
	}

	return summary, nil
}

// PackRecent pages through a pack's successful accesses, newest first.
func (s *PackAccessService) PackRecent(packId uint, q dto.PackAccessRecentQuery) ([]dto.AccessRecord, int64, response.ServerError) {
	return s.recent(SourcePackToml, q, "a.pack_id = ? AND a.success", []any{packId})
}

// SystemRecent pages through the source's records for all packs, newest first,
// optionally filtered by outcome and pack.
func (s *PackAccessService) SystemRecent(source Source, q dto.PackAccessRecentQuery) ([]dto.AccessRecord, int64, response.ServerError) {
	where := "TRUE"
	var args []any
	switch q.Outcome {
	case "success":
		where += " AND a.success"
	case "failure":
		where += " AND NOT a.success"
	}
	if q.PackId != 0 {
		where += " AND a.pack_id = ?"
		args = append(args, q.PackId)
	}
	return s.recent(source, q, where, args)
}

func (s *PackAccessService) recent(source Source, q dto.PackAccessRecentQuery, where string, args []any) ([]dto.AccessRecord, int64, response.ServerError) {
	since := rangeStart(time.Now(), q.Days)
	where += " AND a.created_at >= ?"
	args = append(args, since)

	var total int64
	if err := s.db.Raw("SELECT COUNT(*) FROM "+string(source)+" a WHERE "+where, args...).Scan(&total).Error; err != nil {
		return nil, 0, response.Wrap(fmt.Errorf("pack access count: %w", err))
	}

	records := []dto.AccessRecord{}
	pageArgs := append(append([]any{}, args...), q.PageSize, (q.Page-1)*q.PageSize)
	err := s.db.Raw(`
		SELECT a.id, to_char(a.created_at AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"') AS created_at,
		       a.pack_id, a.pack_slug, COALESCE(p.name, '') AS pack_name,
		       a.user_id, COALESCE(u.username, '') AS username,
		       a.ip_address, a.user_agent, a.status_code, a.success
		FROM `+string(source)+` a
		LEFT JOIN packs p ON p.id = a.pack_id
		LEFT JOIN users u ON u.id = a.user_id
		WHERE `+where+`
		ORDER BY a.created_at DESC, a.id DESC
		LIMIT ? OFFSET ?`, pageArgs...).Scan(&records).Error
	if err != nil {
		return nil, 0, response.Wrap(fmt.Errorf("pack access recent: %w", err))
	}

	return records, total, nil
}

// SystemSummary returns system-wide metrics for the source, including failures.
func (s *PackAccessService) SystemSummary(source Source, days int) (*dto.SystemAccessSummary, response.ServerError) {
	table := string(source)
	now := time.Now()
	since := rangeStart(now, days)
	summary := &dto.SystemAccessSummary{Days: days}

	var sparse []dto.AccessDay
	err := s.db.Raw(`
		SELECT to_char(created_at AT TIME ZONE 'UTC', 'YYYY-MM-DD') AS date,
		       COUNT(*) FILTER (WHERE success) AS success,
		       COUNT(*) FILTER (WHERE NOT success) AS failure
		FROM `+table+`
		WHERE created_at >= ?
		GROUP BY 1`, since).Scan(&sparse).Error
	if err != nil {
		return nil, response.Wrap(fmt.Errorf("system access series: %w", err))
	}
	summary.Series = FillDays(sparse, now, days)

	err = s.db.Raw(`
		SELECT COUNT(*) FILTER (WHERE success),
		       COUNT(*) FILTER (WHERE NOT success),
		       COUNT(DISTINCT ip_address)
		FROM `+table+` WHERE created_at >= ?`, since).
		Row().Scan(&summary.Success, &summary.Failure, &summary.UniqueIps)
	if err != nil {
		return nil, response.Wrap(fmt.Errorf("system access totals: %w", err))
	}

	summary.Packs = []dto.AccessPackCount{}
	err = s.db.Raw(`
		SELECT p.id AS pack_id, p.name, p.slug,
		       COUNT(*) FILTER (WHERE a.success) AS success,
		       COUNT(*) FILTER (WHERE NOT a.success) AS failure
		FROM `+table+` a JOIN packs p ON p.id = a.pack_id
		WHERE a.created_at >= ?
		GROUP BY p.id, p.name, p.slug
		ORDER BY success DESC, failure DESC, p.name`, since).Scan(&summary.Packs).Error
	if err != nil {
		return nil, response.Wrap(fmt.Errorf("system access packs: %w", err))
	}

	summary.TopFailedIp = []dto.AccessIpCount{}
	err = s.db.Raw(`
		SELECT ip_address, COUNT(*) AS count
		FROM `+table+`
		WHERE NOT success AND created_at >= ?
		GROUP BY ip_address
		ORDER BY count DESC, ip_address LIMIT ?`, since, topLimit).Scan(&summary.TopFailedIp).Error
	if err != nil {
		return nil, response.Wrap(fmt.Errorf("system access failed ips: %w", err))
	}

	return summary, nil
}
