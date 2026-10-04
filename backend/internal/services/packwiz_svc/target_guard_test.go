package packwiz_svc

import (
	"net/http"
	"testing"

	"packwiz-web/internal/types"
	"packwiz-web/internal/types/response"
)

func TestCheckPublishedTargetChange(t *testing.T) {
	published, draft := types.PackStatusPublished, types.PackStatusDraft
	for _, tc := range []struct {
		name                         string
		status                       types.PackStatus
		fromMC, fromLoader, toMC, to string
		blocked                      bool
	}{
		{"same target", published, "1.21.1", "fabric", "1.21.1", "fabric", false},
		{"mc upgrade", published, "1.21.1", "fabric", "26.1", "fabric", false},
		{"mc downgrade", published, "26.1", "fabric", "1.21.1", "fabric", true},
		{"downgrade to dev build", published, "26.3", "fabric", "26.3-rc-1", "fabric", true},
		{"loader change", published, "1.21.1", "fabric", "1.21.1", "neoforge", true},
		{"loader case only", published, "1.21.1", "Fabric", "1.21.1", "fabric", false},
		{"weekly snapshot let through", published, "1.21.1", "fabric", "24w14a", "fabric", false},
		{"draft loader change", draft, "1.21.1", "fabric", "1.20.1", "forge", false},
	} {
		err := checkPublishedTargetChange(tc.status, tc.fromMC, tc.fromLoader, tc.toMC, tc.to)
		if (err != nil) != tc.blocked {
			t.Errorf("%s: err = %v, blocked want %v", tc.name, err, tc.blocked)
			continue
		}
		if err != nil && err.(*response.HttpError).Code != http.StatusConflict {
			t.Errorf("%s: code = %d, want 409", tc.name, err.(*response.HttpError).Code)
		}
	}
}
