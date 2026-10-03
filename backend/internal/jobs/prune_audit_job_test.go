package jobs

import (
	"context"
	"testing"
	"time"

	"packwiz-web/internal/config"
)

type fakePruner struct {
	calls  int
	cutoff time.Time
}

func (f *fakePruner) PruneOlderThan(_ context.Context, cutoff time.Time) (int64, error) {
	f.calls++
	f.cutoff = cutoff
	return 3, nil
}

func TestPruneAuditWorkerUsesRetention(t *testing.T) {
	old := config.C.AuditRetentionDays
	defer func() { config.C.AuditRetentionDays = old }()

	config.C.AuditRetentionDays = 90
	p := &fakePruner{}
	w := &PruneAuditWorker{pruner: p}
	if err := w.Work(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
	if p.calls != 1 {
		t.Fatalf("calls = %d, want 1", p.calls)
	}
	want := time.Now().AddDate(0, 0, -90)
	if d := p.cutoff.Sub(want); d < -time.Minute || d > time.Minute {
		t.Errorf("cutoff = %v, want about %v", p.cutoff, want)
	}
}

func TestPruneAuditDisabledWhenZero(t *testing.T) {
	old := config.C.AuditRetentionDays
	defer func() { config.C.AuditRetentionDays = old }()

	config.C.AuditRetentionDays = 0
	p := &fakePruner{}
	if err := (&PruneAuditWorker{pruner: p}).Work(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
	if p.calls != 0 {
		t.Errorf("pruner called %d times with retention 0", p.calls)
	}
	if jobs := auditPrunePeriodicJobs(); jobs != nil {
		t.Errorf("periodic jobs = %v, want none", jobs)
	}
}
