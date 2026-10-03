package jobs

import (
	"context"
	"testing"
	"time"

	"packwiz-web/internal/config"
)

func TestPrunePackAccessWorkerUsesRetention(t *testing.T) {
	old := config.C.AuditRetentionDays
	defer func() { config.C.AuditRetentionDays = old }()

	config.C.AuditRetentionDays = 30
	p := &fakePruner{}
	if err := (&PrunePackAccessWorker{pruner: p}).Work(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
	if p.calls != 1 {
		t.Fatalf("calls = %d, want 1", p.calls)
	}
	want := time.Now().AddDate(0, 0, -30)
	if d := p.cutoff.Sub(want); d < -time.Minute || d > time.Minute {
		t.Errorf("cutoff = %v, want about %v", p.cutoff, want)
	}
}

func TestPrunePackAccessDisabledWhenZero(t *testing.T) {
	old := config.C.AuditRetentionDays
	defer func() { config.C.AuditRetentionDays = old }()

	config.C.AuditRetentionDays = 0
	p := &fakePruner{}
	if err := (&PrunePackAccessWorker{pruner: p}).Work(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
	if p.calls != 0 {
		t.Errorf("pruner called %d times with retention 0", p.calls)
	}
	if jobs := packAccessPrunePeriodicJobs(); jobs != nil {
		t.Errorf("periodic jobs = %v, want none", jobs)
	}
}
