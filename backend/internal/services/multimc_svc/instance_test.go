package multimc_svc

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"testing"
)

func fabricInstance() Instance {
	return Instance{
		Name:          "blarg",
		Description:   "line one\nline two",
		Author:        "leo",
		Version:       "1.2.0",
		MCVersion:     "1.21.1",
		Loader:        "fabric",
		LoaderVersion: "0.19.5",
		PackURL:       "https://pw.example.com/packwiz/tok/blarg/pack.toml",
	}
}

func TestBuildMMCPackFabric(t *testing.T) {
	pack, err := buildMMCPack(fabricInstance())
	if err != nil {
		t.Fatal(err)
	}

	want := map[string]string{
		"net.minecraft":              "1.21.1",
		"net.fabricmc.intermediary":  "1.21.1",
		"net.fabricmc.fabric-loader": "0.19.5",
	}
	if len(pack.Components) != len(want) {
		t.Fatalf("got %d components, want %d", len(pack.Components), len(want))
	}
	for _, c := range pack.Components {
		if want[c.UID] != c.Version {
			t.Errorf("%s: version %q, want %q", c.UID, c.Version, want[c.UID])
		}
	}
	if pack.FormatVersion != 1 {
		t.Errorf("formatVersion = %d", pack.FormatVersion)
	}
}

func TestBuildMMCPackLoaders(t *testing.T) {
	cases := map[string]struct {
		uid          string
		intermediary bool
	}{
		"quilt":    {"org.quiltmc.quilt-loader", true},
		"forge":    {"net.minecraftforge", false},
		"neoforge": {"net.neoforged", false},
	}
	for loader, tc := range cases {
		i := fabricInstance()
		i.Loader = loader
		pack, err := buildMMCPack(i)
		if err != nil {
			t.Fatalf("%s: %v", loader, err)
		}

		var hasLoader, hasInter bool
		for _, c := range pack.Components {
			hasLoader = hasLoader || c.UID == tc.uid
			hasInter = hasInter || c.UID == uidIntermediary
		}
		if !hasLoader || hasInter != tc.intermediary {
			t.Errorf("%s: loader=%v intermediary=%v, components=%+v", loader, hasLoader, hasInter, pack.Components)
		}
	}
}

func TestBuildMMCPackErrors(t *testing.T) {
	noMC := fabricInstance()
	noMC.MCVersion = ""
	badLoader := fabricInstance()
	badLoader.Loader = "nope"
	noLoaderVer := fabricInstance()
	noLoaderVer.LoaderVersion = ""

	for name, i := range map[string]Instance{"no mc": noMC, "bad loader": badLoader, "no loader version": noLoaderVer} {
		if _, err := buildMMCPack(i); err == nil {
			t.Errorf("%s: expected error", name)
		}
	}
}

func TestBuildMMCPackVanilla(t *testing.T) {
	i := fabricInstance()
	i.Loader, i.LoaderVersion = "", ""
	pack, err := buildMMCPack(i)
	if err != nil {
		t.Fatal(err)
	}
	if len(pack.Components) != 1 || pack.Components[0].UID != uidMinecraft {
		t.Errorf("components = %+v", pack.Components)
	}
}

func TestBuildInstanceCfg(t *testing.T) {
	cfg := buildInstanceCfg(fabricInstance())

	for _, want := range []string{
		"name=blarg\n",
		"MaxMemAlloc=4096\n",
		"MinMemAlloc=2048\n",
		"JvmArgs=-XX:+UseZGC -XX:+ZGenerational\n",
		"OverrideCommands=true\n",
		`PreLaunchCommand=\"$INST_JAVA\" -jar packwiz-installer-bootstrap.jar https://pw.example.com/packwiz/tok/blarg/pack.toml` + "\n",
	} {
		if !strings.Contains(cfg, want) {
			t.Errorf("instance.cfg missing %q\n%s", want, cfg)
		}
	}

	// notes must stay on one line, with the description, metadata and source in it
	var notes string
	for _, line := range strings.Split(cfg, "\n") {
		if strings.HasPrefix(line, "notes=") {
			notes = line
		}
	}
	for _, want := range []string{
		`line one\nline two`,
		`Author: leo`,
		`Pack version: 1.2.0`,
		`Minecraft: 1.21.1`,
		`Loader: fabric 0.19.5`,
		`Source: https://pw.example.com/packwiz/tok/blarg/pack.toml`,
	} {
		if !strings.Contains(notes, want) {
			t.Errorf("notes missing %q: %s", want, notes)
		}
	}
}

// parseLegacyCfg ports the line parser that upstream MultiMC and Prism (for
// files with no ConfigVersion) use: strip an unescaped "#" comment, split on
// the first "=", trim, then unescape.
func parseLegacyCfg(cfg string) map[string]string {
	out := map[string]string{}
	for _, line := range strings.Split(cfg, "\n") {
		for i := 1; i < len(line); i++ {
			if line[i] == '#' && line[i-1] != '\\' {
				line = strings.TrimSpace(line[:i])
				break
			}
		}
		eq := strings.Index(line, "=")
		if eq == -1 {
			continue
		}
		key := strings.TrimSpace(line[:eq])
		raw := strings.TrimSpace(line[eq+1:])

		var b strings.Builder
		escaped := false
		for _, c := range raw {
			switch {
			case escaped && c == 'n':
				b.WriteRune('\n')
			case escaped && c == 't':
				b.WriteRune('\t')
			case escaped:
				b.WriteRune(c)
			case c == '\\':
				escaped = true
				continue
			default:
				b.WriteRune(c)
			}
			escaped = false
		}
		out[key] = b.String()
	}
	return out
}

func TestInstanceCfgRoundTrip(t *testing.T) {
	i := fabricInstance()
	i.Name = "Pack, the; #1 \"best\""
	i.Description = "# Heading\r\nmods, configs; and more\ttabbed\\path\nlast = line"

	got := parseLegacyCfg(buildInstanceCfg(i))

	if got["name"] != i.Name {
		t.Errorf("name = %q, want %q", got["name"], i.Name)
	}
	wantNotes := strings.ReplaceAll(instanceNotes(i), "\r\n", "\n")
	if got["notes"] != wantNotes {
		t.Errorf("notes = %q, want %q", got["notes"], wantNotes)
	}
	wantPre := `"$INST_JAVA" -jar packwiz-installer-bootstrap.jar ` + i.PackURL
	if got["PreLaunchCommand"] != wantPre {
		t.Errorf("PreLaunchCommand = %q, want %q", got["PreLaunchCommand"], wantPre)
	}
	if got["InstanceType"] != "OneSix" {
		t.Errorf("InstanceType = %q", got["InstanceType"])
	}
}

func TestInstanceCfgIsLegacyFormat(t *testing.T) {
	cfg := buildInstanceCfg(fabricInstance())
	// ConfigVersion makes Prism parse with QSettings, which mangles "," and ";"
	if strings.Contains(cfg, "ConfigVersion") || strings.Contains(cfg, "[General]") {
		t.Errorf("instance.cfg must use the legacy format:\n%s", cfg)
	}
}

func TestBuildZip(t *testing.T) {
	data, err := BuildZip(fabricInstance())
	if err != nil {
		t.Fatal(err)
	}

	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatal(err)
	}

	files := map[string][]byte{}
	for _, f := range zr.File {
		rc, err := f.Open()
		if err != nil {
			t.Fatal(err)
		}
		files[f.Name], _ = io.ReadAll(rc)
		rc.Close()
	}

	for _, name := range []string{"instance.cfg", "mmc-pack.json", "minecraft/packwiz-installer-bootstrap.jar"} {
		if len(files[name]) == 0 {
			t.Errorf("zip missing or empty: %s", name)
		}
	}

	var pack mmcPack
	if err := json.Unmarshal(files["mmc-pack.json"], &pack); err != nil {
		t.Fatalf("mmc-pack.json invalid: %v", err)
	}
	if !bytes.HasPrefix(files["minecraft/packwiz-installer-bootstrap.jar"], []byte("PK")) {
		t.Error("bootstrap jar is not a zip/jar")
	}
}

func TestRequiresJava21(t *testing.T) {
	for v, want := range map[string]bool{
		"1.16.5": false, "1.20.1": false, "1.20.4": false, "1.20": false,
		"1.20.5": true, "1.20.6": true, "1.21": true, "1.21.1": true,
		"26.1": true, "24w14a": false, "": false,
	} {
		if got := requiresJava21(v); got != want {
			t.Errorf("requiresJava21(%q) = %v, want %v", v, got, want)
		}
	}
}

func TestInstanceCfgJvmArgsGatedOnJava(t *testing.T) {
	old := fabricInstance()
	old.MCVersion = "1.20.1"
	cfg := buildInstanceCfg(old)
	if !strings.Contains(cfg, "JvmArgs=\n") || !strings.Contains(cfg, "OverrideJavaArgs=false\n") {
		t.Errorf("1.20.1 must not set ZGC args:\n%s", cfg)
	}
	if cfg := buildInstanceCfg(fabricInstance()); !strings.Contains(cfg, "OverrideJavaArgs=true\n") {
		t.Errorf("1.21.1 should set ZGC args:\n%s", cfg)
	}
}
