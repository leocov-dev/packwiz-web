package packwiz_svc

import (
	"testing"

	"packwiz-web/internal/tables"
	"packwiz-web/internal/types/dto"
)

func TestApplyCompat(t *testing.T) {
	packs := []dto.PackResponse{{Pack: tables.Pack{ID: 1}}, {Pack: tables.Pack{ID: 2}}}
	applyCompat(packs, map[uint]dto.CompatSummary{2: {Status: "errors", Errors: 3}})
	if packs[0].Compat != nil {
		t.Error("pack 1 never scanned: want nil")
	}
	if packs[1].Compat == nil || packs[1].Compat.Errors != 3 {
		t.Errorf("pack 2: %+v", packs[1].Compat)
	}
}

func TestAttachCompatNilSafe(t *testing.T) {
	ps := &PackwizService{}
	ps.attachCompat([]dto.PackResponse{{Pack: tables.Pack{ID: 1}}}) // no summarizer: no-op
}
