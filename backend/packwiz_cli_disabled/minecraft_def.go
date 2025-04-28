package packwiz_cli_disabled

// MinecraftDef
// when Latest is True the Version is ignored
// when Latest is True, the latest Release version will be used unless Snapshot is True, then the latest Snapshot version will be used
type MinecraftDef struct {
	Version  string `json:"version"`
	Latest   bool   `json:"latest"`
	Snapshot bool   `json:"snapshot"`
}

func (m MinecraftDef) AsArgs() []string {
	args := make([]string, 0)

	if m.Version != "" {
		args = append(args, "--mc-version", m.Version)
	} else {
		args = append(args, "--latest")
	}

	if m.Snapshot {
		args = append(args, "--snapshot")
	}

	return args
}
