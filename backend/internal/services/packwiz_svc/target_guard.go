package packwiz_svc

import (
	"fmt"
	"net/http"
	"strings"

	"packwiz-web/internal/types"
	"packwiz-web/internal/types/response"
	"packwiz-web/internal/utils"
)

// checkPublishedTargetChange refuses the Minecraft/loader changes that break
// clients of a published pack: a different loader, or an older Minecraft
// version. packwiz-installer can move an instance to a newer Minecraft or
// loader version, but a loader switch leaves the old loader's mods and configs
// behind, and a downgrade can corrupt worlds. Those need a new pack (clone).
// Drafts are not served to consumers, so they may change freely. A Minecraft
// id that cannot be compared (e.g. a weekly snapshot) is let through.
func checkPublishedTargetChange(status types.PackStatus, fromMC, fromLoader, toMC, toLoader string) response.ServerError {
	if status != types.PackStatusPublished {
		return nil
	}

	if !strings.EqualFold(fromLoader, toLoader) {
		return response.New(http.StatusConflict, fmt.Sprintf(
			"a published pack cannot change its loader (%s to %s); clone it to a new pack instead",
			fromLoader, toLoader,
		))
	}

	if cmp, ok := utils.CompareMinecraftVersions(toMC, fromMC); ok && cmp < 0 {
		return response.New(http.StatusConflict, fmt.Sprintf(
			"a published pack cannot move to an older Minecraft version (%s to %s); clone it to a new pack instead",
			fromMC, toMC,
		))
	}

	return nil
}
