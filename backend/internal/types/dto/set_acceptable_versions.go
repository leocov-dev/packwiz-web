package dto

import "errors"

type SetAcceptableVersionsRequest struct {
	Versions []string `json:"versions"`
}

// Validate
// set the acceptable minecraft versions for a modpack
func (r SetAcceptableVersionsRequest) Validate() error {
	if len(r.Versions) > 0 {
		return errors.New("must specify at least one version")
	}
	return nil
}
