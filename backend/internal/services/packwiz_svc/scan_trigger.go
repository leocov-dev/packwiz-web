package packwiz_svc

import (
	"context"
	"errors"
	"fmt"

	"packwiz-web/internal/jobs"
	"packwiz-web/internal/log"
)

// ScanRequester schedules a background jar compatibility scan of a pack. It is
// implemented by compat_svc.CompatService. packwiz_svc depends on this small
// interface instead of the compat service, so the two services stay decoupled
// (compat_svc must not import packwiz_svc, and the interface keeps it
// substitutable in tests).
//
// Implementations debounce: repeated requests for a pack collapse into one
// scheduled scan.
type ScanRequester interface {
	RequestScan(ctx context.Context, packID uint) error
}

// requestCompatScan asks for a scan of the pack. Call it after the change has
// committed, never inside its transaction. A failure to enqueue must not fail
// the user's operation, so it is only logged. A nil requester does nothing.
func (ps *PackwizService) requestCompatScan(ctx context.Context, packID uint) {
	if ps.scans == nil {
		return
	}
	if err := ps.scans.RequestScan(ctx, packID); err != nil {
		// instances without a job client (the worker's job resolver) report
		// that every time; their jobs request the scan with their own context
		if errors.Is(err, jobs.ErrClientUnavailable) {
			log.Debug(fmt.Sprintf("compat scan of pack %d not requested:", packID), err)
			return
		}
		log.Error(fmt.Errorf("request compat scan of pack %d: %w", packID, err))
	}
}
