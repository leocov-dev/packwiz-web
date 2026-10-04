package dto

// PackLinksResponse holds a user's consumer links for one pack.
type PackLinksResponse struct {
	// Link is the pack.toml link for packwiz installers.
	Link string `json:"link"`
	// MultiMCLink is the importable MultiMC / Prism instance zip.
	MultiMCLink string `json:"multimcLink"`
}
