package packwiz_svc

import (
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"gorm.io/gorm"

	"packwiz-web/internal/tables"
	"packwiz-web/internal/types/dto"
	"packwiz-web/internal/types/response"
)

// errRemoveBlocked marks a removal that would leave a mod without a
// dependency it requires.
var errRemoveBlocked = errors.New("remove blocked")

// blockedRemovals lists the mods to remove that are dependencies still
// required by a mod staying in the pack, as "Dep (required by A, B)", sorted.
func blockedRemovals(mods []tables.Mod, remove map[uint]bool) []string {
	names := make(map[uint]string, len(mods))
	for _, m := range mods {
		names[m.ID] = m.Name
	}
	requiredBy := map[uint][]string{}
	for _, m := range mods {
		if remove[m.ID] {
			continue
		}
		for _, depId := range m.DependencyIds {
			if depId != m.ID && remove[depId] {
				requiredBy[depId] = append(requiredBy[depId], m.Name)
			}
		}
	}
	out := make([]string, 0, len(requiredBy))
	for depId, by := range requiredBy {
		sort.Strings(by)
		out = append(out, fmt.Sprintf("%s (required by %s)", names[depId], strings.Join(by, ", ")))
	}
	sort.Strings(out)
	return out
}

// RemoveMods removes several mods of a pack in one history snapshot. All ids
// must belong to the pack (404 otherwise). A dependency can only be removed
// together with every mod that still requires it (409 otherwise), as in the UI.
func (ps *PackwizService) RemoveMods(packId uint, request dto.RemoveModsRequest, user tables.User) (dto.RemoveModsResponse, response.ServerError) {
	remove := make(map[uint]bool, len(request.ModIds))
	for _, id := range request.ModIds {
		remove[id] = true
	}

	var blocked []string
	detail := map[string]any{}
	if err := ps.withPackHistory(packId, user.ID, tables.SnapshotModRemove, detail, func(tx *gorm.DB) error {
		var mods []tables.Mod
		if err := tx.Select("id", "slug", "name", "dependency_ids").
			Where("pack_id = ?", packId).Find(&mods).Error; err != nil {
			return err
		}

		var names []string
		for _, m := range mods {
			if remove[m.ID] {
				names = append(names, m.Name)
			}
		}
		if len(names) != len(remove) {
			return gorm.ErrRecordNotFound
		}
		if blocked = blockedRemovals(mods, remove); len(blocked) > 0 {
			return errRemoveBlocked
		}

		sort.Strings(names)
		detail["count"] = len(names)
		detail["names"] = names
		return tx.Where("pack_id = ? AND id IN ?", packId, request.ModIds).Delete(&tables.Mod{}).Error
	}); err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return dto.RemoveModsResponse{}, response.New(http.StatusNotFound, fmt.Sprintf("some mods are not in pack %d", packId))
		case errors.Is(err, errRemoveBlocked):
			return dto.RemoveModsResponse{}, response.New(http.StatusConflict, "still required: "+strings.Join(blocked, "; "))
		}
		return dto.RemoveModsResponse{}, response.Wrap(err)
	}

	return dto.RemoveModsResponse{Removed: len(remove)}, nil
}
