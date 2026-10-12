package compat_svc

import (
	"errors"
	"time"
)

const (
	// ScanDownloadTimeBudget bounds the time one scan spends downloading jars.
	// It is below jobs.CompatScanTimeout (30 minutes) by more than one
	// DownloadTimeout, so the download that crosses the budget still finishes
	// inside the job timeout.
	ScanDownloadTimeBudget = 24 * time.Minute
	// ScanDownloadByteBudget bounds the bytes one scan downloads.
	ScanDownloadByteBudget int64 = 1 << 30
)

// errBudgetExhausted means the scan used up its download budget; the mod is
// counted as unscanned and the run is left for the next catch-up (see scan).
var errBudgetExhausted = errors.New("scan download budget exhausted")

// downloadBudget tracks what one scan may still download. Not safe for
// concurrent use: a scan resolves jars one at a time.
type downloadBudget struct {
	timeLeft  time.Duration
	bytesLeft int64
	// exhausted is set once a download was refused for lack of budget.
	exhausted bool
}

func newDownloadBudget() *downloadBudget {
	return &downloadBudget{timeLeft: ScanDownloadTimeBudget, bytesLeft: ScanDownloadByteBudget}
}

// allow reports whether another download may start. A refusal is remembered.
func (b *downloadBudget) allow() bool {
	if b.timeLeft <= 0 || b.bytesLeft <= 0 {
		b.exhausted = true
		return false
	}
	return true
}

// spend records a finished download attempt.
func (b *downloadBudget) spend(d time.Duration, bytes int64) {
	b.timeLeft -= d
	b.bytesLeft -= bytes
}
