package jobs

import (
	"context"

	"github.com/riverqueue/river"
)

// QueueMigrateMods is a dedicated queue (capped at 1 worker, see client.go)
// for MigrateModsArgs jobs, kept separate from river.QueueDefault until the
// packwiz-nxt source Updaters are verified safe for concurrent use.
const QueueMigrateMods = "migrate_mods"

// MigrateModsArgs re-checks and, where possible, updates a pack's mods
// against the Minecraft version / loader target that PackwizService.Migrate
// already persisted before enqueueing this job.
type MigrateModsArgs struct {
	PackID             uint     `json:"packId"`
	MCVersion          string   `json:"mcVersion"`
	LoaderName         string   `json:"loaderName"`
	LoaderVersion      string   `json:"loaderVersion"`
	AcceptableVersions []string `json:"acceptableVersions"`
	UserID             uint     `json:"userId"`
}

func (MigrateModsArgs) Kind() string { return "migrate_mods" }

func (MigrateModsArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{Queue: QueueMigrateMods}
}

// MigrateModsResolver is implemented by PackwizService. It's expressed as an
// interface here (rather than importing packwiz_svc directly) to avoid an
// import cycle: packwiz_svc imports jobs (to build MigrateModsArgs and enqueue
// them), so jobs cannot import packwiz_svc back.
type MigrateModsResolver interface {
	ResolveMigratedMods(ctx context.Context, args MigrateModsArgs, jobId int64) error
}

type MigrateModsWorker struct {
	river.WorkerDefaults[MigrateModsArgs]
	resolver MigrateModsResolver
}

func (w *MigrateModsWorker) Work(ctx context.Context, job *river.Job[MigrateModsArgs]) error {
	return w.resolver.ResolveMigratedMods(ctx, job.Args, job.ID)
}
