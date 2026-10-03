package pack_access_svc

import (
	"time"

	"packwiz-web/internal/types/dto"
)

const dateLayout = "2006-01-02"

// rangeStart is the start (UTC midnight) of the first day of a trailing
// window of the given number of days that ends today.
func rangeStart(now time.Time, days int) time.Time {
	today := now.UTC().Truncate(24 * time.Hour)
	return today.AddDate(0, 0, -(days - 1))
}

// FillDays returns one entry per UTC day of the trailing window, oldest first,
// taking counts from sparse (keyed by YYYY-MM-DD) and zero-filling the rest.
func FillDays(sparse []dto.AccessDay, now time.Time, days int) []dto.AccessDay {
	byDate := make(map[string]dto.AccessDay, len(sparse))
	for _, d := range sparse {
		byDate[d.Date] = d
	}

	start := rangeStart(now, days)
	out := make([]dto.AccessDay, 0, days)
	for i := 0; i < days; i++ {
		date := start.AddDate(0, 0, i).Format(dateLayout)
		day, ok := byDate[date]
		if !ok {
			day = dto.AccessDay{Date: date}
		}
		out = append(out, day)
	}
	return out
}
