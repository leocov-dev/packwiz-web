package packwiz_cli_disabled

type LoaderType string

const (
	FabricLoader   LoaderType = "fabric"
	ForgeLoader    LoaderType = "forge"
	LiteLoader     LoaderType = "liteloader"
	QuiltLoader    LoaderType = "quilt"
	NeoForgeLoader LoaderType = "neoforge"
)

// LoaderDef
// when Latest is True the Version is ignored
type LoaderDef struct {
	Name    LoaderType `json:"name"`
	Version string     `json:"version"`
	Latest  bool       `json:"latest"`
}

func (l LoaderDef) AsArgs() []string {
	args := []string{
		"--modloader", string(l.Name),
	}

	if l.Name == FabricLoader {
		if l.Version != "" {
			args = append(args, "--fabric-version", l.Version)
		} else {
			args = append(args, "--fabric-latest")
		}
	}

	if l.Name == ForgeLoader {
		if l.Version != "" {
			args = append(args, "--forge-version", l.Version)
		} else {
			args = append(args, "--forge-latest")
		}
	}

	if l.Name == LiteLoader {
		if l.Version != "" {
			args = append(args, "--liteloader-version", l.Version)
		} else {
			args = append(args, "--liteloader-latest")
		}
	}

	if l.Name == QuiltLoader {
		if l.Version != "" {
			args = append(args, "--quilt-version", l.Version)
		} else {
			args = append(args, "--quilt-latest")
		}
	}

	if l.Name == NeoForgeLoader {
		if l.Version != "" {
			args = append(args, "--neoforge-version", l.Version)
		} else {
			args = append(args, "--neoforge-latest")
		}
	}

	return args
}
