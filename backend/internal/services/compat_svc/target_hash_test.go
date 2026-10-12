package compat_svc

import "testing"

func TestTargetHash(t *testing.T) {
	base := TargetInput{
		MinecraftVersion: "1.21.1",
		Loader:           "fabric",
		LoaderVersion:    "0.16.0",
		Mods: []HashedMod{
			{Slug: "a", HashFormat: "sha1", Hash: "h1", Side: "both"},
			{Slug: "b", HashFormat: "sha512", Hash: "h2", Side: "both"},
		},
	}
	h := TargetHash(base)

	if h != TargetHash(base) {
		t.Fatal("hash not stable")
	}

	reordered := base
	reordered.Mods = []HashedMod{base.Mods[1], base.Mods[0]}
	if TargetHash(reordered) != h {
		t.Error("hash depends on mod order")
	}

	tests := []struct {
		name   string
		mutate func(in *TargetInput)
	}{
		{"mc version", func(in *TargetInput) { in.MinecraftVersion = "1.21.2" }},
		{"loader", func(in *TargetInput) { in.Loader = "quilt" }},
		{"loader version", func(in *TargetInput) { in.LoaderVersion = "0.16.1" }},
		{"slug", func(in *TargetInput) { in.Mods = []HashedMod{{"x", "sha1", "h1", "both"}, base.Mods[1]} }},
		{"hash format", func(in *TargetInput) { in.Mods = []HashedMod{{"a", "md5", "h1", "both"}, base.Mods[1]} }},
		{"hash", func(in *TargetInput) { in.Mods = []HashedMod{{"a", "sha1", "zz", "both"}, base.Mods[1]} }},
		{"side", func(in *TargetInput) { in.Mods = []HashedMod{{"a", "sha1", "h1", "server"}, base.Mods[1]} }},
		{"mod removed", func(in *TargetInput) { in.Mods = base.Mods[:1] }},
		{"field boundary", func(in *TargetInput) { in.MinecraftVersion, in.Loader = "1.21.1f", "abric" }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := base
			in.Mods = append([]HashedMod(nil), base.Mods...)
			tt.mutate(&in)
			if TargetHash(in) == h {
				t.Error("hash did not change")
			}
		})
	}
}
