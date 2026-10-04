package multimc_svc

import (
	"archive/zip"
	"bytes"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// bootstrapJar is packwiz-installer-bootstrap (MIT, v0.0.3), vendored so an
// imported instance can sync itself on first launch without a manual download.
// See assets/README.md.
//
//go:embed assets/packwiz-installer-bootstrap.jar
var bootstrapJar []byte

const (
	bootstrapJarName = "packwiz-installer-bootstrap.jar"

	// MultiMC treats "minecraft" and ".minecraft" the same on import.
	gameDir = "minecraft"

	// memory is capped at 4 GiB on purpose; users can raise it in the client.
	minMemoryMB = 2048
	maxMemoryMB = 4096

	// zgcArgs need Java 21, which Minecraft only requires from 1.20.5. Older
	// JVMs refuse to start on the unknown flag, so they are gated on the version.
	zgcArgs = "-XX:+UseZGC -XX:+ZGenerational"
)

// requiresJava21 reports whether an MC release needs Java 21 or newer (1.20.5+,
// and the 26.x year-based releases). Snapshots and anything unparseable return
// false, which just means no custom JVM args.
func requiresJava21(mcVersion string) bool {
	parts := strings.Split(mcVersion, ".")
	nums := make([]int, 0, 3)
	for _, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil {
			return false
		}
		nums = append(nums, n)
	}
	if len(nums) < 2 {
		return false
	}

	if nums[0] != 1 {
		return nums[0] >= 26
	}
	if nums[1] != 20 {
		return nums[1] > 20
	}
	return len(nums) > 2 && nums[2] >= 5
}

// ErrInvalidPack marks a pack that cannot be described as an instance, like one
// with no Minecraft version or an unknown loader.
var ErrInvalidPack = errors.New("pack cannot be exported as an instance")

// Instance is everything needed to describe a pack as a MultiMC instance.
type Instance struct {
	Name          string
	Description   string
	Author        string
	Version       string
	MCVersion     string
	Loader        string
	LoaderVersion string
	// PackURL is the pack.toml link the installer syncs from (personal or public).
	PackURL string
}

type mmcComponent struct {
	CachedName     string       `json:"cachedName,omitempty"`
	CachedRequires []mmcRequire `json:"cachedRequires,omitempty"`
	CachedVersion  string       `json:"cachedVersion,omitempty"`
	CachedVolatile bool         `json:"cachedVolatile,omitempty"`
	DependencyOnly bool         `json:"dependencyOnly,omitempty"`
	Important      bool         `json:"important,omitempty"`
	UID            string       `json:"uid"`
	Version        string       `json:"version"`
}

type mmcRequire struct {
	Equals string `json:"equals,omitempty"`
	UID    string `json:"uid"`
}

type mmcPack struct {
	Components    []mmcComponent `json:"components"`
	FormatVersion int            `json:"formatVersion"`
}

const (
	uidMinecraft    = "net.minecraft"
	uidIntermediary = "net.fabricmc.intermediary"
)

// loaderComponents maps a packwiz loader name to its MultiMC component uid and
// display name. Fabric and Quilt both sit on the fabric intermediary mappings.
var loaderComponents = map[string]struct {
	uid, name    string
	intermediary bool
}{
	"fabric":     {"net.fabricmc.fabric-loader", "Fabric Loader", true},
	"quilt":      {"org.quiltmc.quilt-loader", "Quilt Loader", true},
	"forge":      {"net.minecraftforge", "Forge", false},
	"neoforge":   {"net.neoforged", "NeoForge", false},
	"liteloader": {"com.mumfrey.liteloader", "LiteLoader", false},
}

// buildMMCPack returns the component list for mmc-pack.json. LWJGL is left out:
// MultiMC and Prism resolve it from the Minecraft component's requirements, so
// we don't have to track which LWJGL each Minecraft version needs.
func buildMMCPack(i Instance) (mmcPack, error) {
	if i.MCVersion == "" {
		return mmcPack{}, fmt.Errorf("%w: pack has no minecraft version", ErrInvalidPack)
	}

	components := []mmcComponent{{
		CachedName:    "Minecraft",
		CachedVersion: i.MCVersion,
		Important:     true,
		UID:           uidMinecraft,
		Version:       i.MCVersion,
	}}

	loader := strings.ToLower(i.Loader)
	if loader != "" {
		def, ok := loaderComponents[loader]
		if !ok {
			return mmcPack{}, fmt.Errorf("%w: unsupported loader %q", ErrInvalidPack, i.Loader)
		}
		if i.LoaderVersion == "" {
			return mmcPack{}, fmt.Errorf("%w: pack has no %s version", ErrInvalidPack, loader)
		}

		if def.intermediary {
			components = append(components, mmcComponent{
				CachedName:     "Intermediary Mappings",
				CachedRequires: []mmcRequire{{Equals: i.MCVersion, UID: uidMinecraft}},
				CachedVersion:  i.MCVersion,
				CachedVolatile: true,
				DependencyOnly: true,
				UID:            uidIntermediary,
				Version:        i.MCVersion,
			})
		}

		components = append(components, mmcComponent{
			CachedName:    def.name,
			CachedVersion: i.LoaderVersion,
			UID:           def.uid,
			Version:       i.LoaderVersion,
		})
	}

	return mmcPack{Components: components, FormatVersion: 1}, nil
}

// cfgValue escapes a value for one instance.cfg line in MultiMC's own format:
// "\" escapes, "\n"/"\t" are newline/tab, and an unescaped "#" starts a
// comment. Quotes are escaped as MultiMC writes them; both parsers unescape them.
func cfgValue(s string) string {
	return strings.NewReplacer(
		"\r\n", "\\n",
		"\r", "",
		"\\", "\\\\",
		"\n", "\\n",
		"\t", "\\t",
		"#", "\\#",
		`"`, `\"`,
	).Replace(s)
}

// instanceNotes is the free-form description shown in the client: the pack
// description, its metadata, and where it syncs from.
func instanceNotes(i Instance) string {
	var parts []string
	if d := strings.TrimSpace(i.Description); d != "" {
		parts = append(parts, d)
	}

	var meta []string
	if i.Author != "" {
		meta = append(meta, "Author: "+i.Author)
	}
	if i.Version != "" {
		meta = append(meta, "Pack version: "+i.Version)
	}
	meta = append(meta, "Minecraft: "+i.MCVersion)
	if i.Loader != "" {
		meta = append(meta, fmt.Sprintf("Loader: %s %s", i.Loader, i.LoaderVersion))
	}
	meta = append(meta, "Source: "+i.PackURL)
	parts = append(parts, strings.Join(meta, "\n"))

	return strings.Join(parts, "\n\n")
}

// buildInstanceCfg renders instance.cfg in the legacy MultiMC format: no
// ConfigVersion and no section header. Upstream MultiMC only reads this format.
// Prism reads a file that has ConfigVersion with QSettings, where an unquoted
// "," turns a value into a list and ";" starts a comment, which would empty or
// cut the name and notes. Without it, Prism uses its MultiMC-compatible parser
// and upgrades the file on save. Window/column state is omitted so the client
// applies its own defaults.
func buildInstanceCfg(i Instance) string {
	name := strings.TrimSpace(strings.NewReplacer("\r", " ", "\n", " ").Replace(i.Name))
	preLaunch := fmt.Sprintf(`"$INST_JAVA" -jar %s %s`, bootstrapJarName, i.PackURL)

	javaArgsOverride := "false"
	jvm := ""
	if requiresJava21(i.MCVersion) {
		javaArgsOverride = "true"
		jvm = zgcArgs
	}

	lines := []string{
		"InstanceType=OneSix",
		"JoinServerOnLaunch=false",
		"JvmArgs=" + jvm,
		fmt.Sprintf("MaxMemAlloc=%d", maxMemoryMB),
		fmt.Sprintf("MinMemAlloc=%d", minMemoryMB),
		"OverrideCommands=true",
		"OverrideConsole=false",
		"OverrideEnv=false",
		"OverrideGameTime=false",
		"OverrideJavaArgs=" + javaArgsOverride,
		"OverrideJavaLocation=false",
		"OverrideLegacySettings=false",
		"OverrideMemory=true",
		"OverrideMiscellaneous=false",
		"OverrideNativeWorkarounds=false",
		"OverridePerformance=false",
		"OverrideWindow=false",
		"PermGen=128",
		"PostExitCommand=",
		"PreLaunchCommand=" + cfgValue(preLaunch),
		"UseAccountForInstance=false",
		"WrapperCommand=",
		"iconKey=default",
		"name=" + cfgValue(name),
		"notes=" + cfgValue(instanceNotes(i)),
		"",
	}

	return strings.Join(lines, "\n")
}

// BuildZip assembles the importable instance archive: instance.cfg,
// mmc-pack.json, and the packwiz installer in the game directory.
func BuildZip(i Instance) ([]byte, error) {
	pack, err := buildMMCPack(i)
	if err != nil {
		return nil, err
	}

	packJSON, err := json.MarshalIndent(pack, "", "    ")
	if err != nil {
		return nil, fmt.Errorf("encode mmc-pack.json: %w", err)
	}

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	files := []struct {
		name string
		data []byte
	}{
		{"instance.cfg", []byte(buildInstanceCfg(i))},
		{"mmc-pack.json", packJSON},
		{gameDir + "/" + bootstrapJarName, bootstrapJar},
	}
	for _, f := range files {
		w, err := zw.Create(f.name)
		if err != nil {
			return nil, fmt.Errorf("zip %s: %w", f.name, err)
		}
		if _, err := w.Write(f.data); err != nil {
			return nil, fmt.Errorf("zip %s: %w", f.name, err)
		}
	}

	if err := zw.Close(); err != nil {
		return nil, fmt.Errorf("finish zip: %w", err)
	}

	return buf.Bytes(), nil
}
