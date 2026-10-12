package packwiz_svc

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"packwiz-web/internal/jobs"
)

type fakeScans struct {
	calls []uint
	err   error
}

func (f *fakeScans) RequestScan(_ context.Context, packID uint) error {
	f.calls = append(f.calls, packID)
	return f.err
}

func TestRequestCompatScan(t *testing.T) {
	tests := []struct {
		name      string
		scans     *fakeScans
		wantCalls int
	}{
		{"requests the scan", &fakeScans{}, 1},
		{"enqueue failure is swallowed", &fakeScans{err: errors.New("db down")}, 1},
		{"no job client is swallowed", &fakeScans{err: fmt.Errorf("x: %w", jobs.ErrClientUnavailable)}, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := &PackwizService{scans: tt.scans}
			ps.requestCompatScan(context.Background(), 9) // must not panic or return an error
			if len(tt.scans.calls) != tt.wantCalls || tt.scans.calls[0] != 9 {
				t.Errorf("calls = %v", tt.scans.calls)
			}
		})
	}

	t.Run("nil requester is a no-op", func(t *testing.T) {
		(&PackwizService{}).requestCompatScan(context.Background(), 9)
	})
}
