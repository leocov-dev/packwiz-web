package packwiz_svc

import (
	"context"
	"errors"
	"fmt"

	"packwiz-web/internal/jobs"
	"packwiz-web/internal/log"
	"packwiz-web/internal/types/dto"
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

// CompatSummarizer loads the jar compatibility badge summaries of many packs in
// one batch. It is implemented by compat_svc.CompatService; like ScanRequester
// it keeps packwiz_svc decoupled from compat_svc. Packs never scanned are
// absent from the result.
type CompatSummarizer interface {
	SummariesForPacks(ctx context.Context, packIDs []uint) (map[uint]dto.CompatSummary, error)
}

// WithCompatSummaries sets the source of the compat badge summaries on pack
// responses and returns the service. Without one (or when it fails) pack
// responses simply carry no summary. Call it right after construction.
func (ps *PackwizService) WithCompatSummaries(summaries CompatSummarizer) *PackwizService {
	ps.compat = summaries
	return ps
}

// attachCompat fills each pack's compat summary with one batched lookup. The
// badge is a nicety, so a failure is logged and the packs are returned as is.
func (ps *PackwizService) attachCompat(packs []dto.PackResponse) {
	if ps.compat == nil || len(packs) == 0 {
		return
	}
	ids := make([]uint, len(packs))
	for i := range packs {
		ids[i] = packs[i].ID
	}
	summaries, err := ps.compat.SummariesForPacks(context.Background(), ids)
	if err != nil {
		if !errors.Is(err, jobs.ErrClientUnavailable) {
			log.Error(fmt.Errorf("load compat summaries: %w", err))
		}
		return
	}
	applyCompat(packs, summaries)
}

// applyCompat sets the summary of each pack that has one.
func applyCompat(packs []dto.PackResponse, summaries map[uint]dto.CompatSummary) {
	for i := range packs {
		if s, ok := summaries[packs[i].ID]; ok {
			packs[i].Compat = &s
		}
	}
}
