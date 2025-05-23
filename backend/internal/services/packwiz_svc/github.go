package packwiz_svc

import (
	"errors"
	"github.com/leocov-dev/packwiz-nxt/core"
	"github.com/leocov-dev/packwiz-nxt/sources"
)

func addGithubMod(url string, _ core.Pack) ([]*core.Mod, error) {

	mod, err := sources.GitHubNewMod(url, "", "", "mods")
	if err != nil {
		return nil, err
	}

	if mod == nil {
		return nil, errors.New("failed to add mod")
	}

	return []*core.Mod{mod}, nil
}
