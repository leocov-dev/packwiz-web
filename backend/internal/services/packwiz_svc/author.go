package packwiz_svc

import (
	"packwiz-web/internal/tables"
)

// GetPackAuthorName returns the username of the pack's author, or "" if it
// cannot be resolved.
func (ps *PackwizService) GetPackAuthorName(packId uint) string {
	var pack tables.Pack
	if err := ps.db.Unscoped().Preload("Author").Select("id", "created_by").First(&pack, packId).Error; err != nil {
		return ""
	}
	return pack.Author.Username
}
