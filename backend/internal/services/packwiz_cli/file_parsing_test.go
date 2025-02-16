package packwiz_cli

import (
	"errors"
	"testing"
)

func TestParsePackFile(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		want    PackFile
		wantErr error
	}{
		{
			name: "valid data",
			data: []byte(`
				name = "ExamplePack"
				version = "1.0.0"
				pack-format = "packwiz:1.1.0"
				
				[index]
				file = "index.toml"
				hash-format = "sha256"
				hash = "e2f46ef6ea8628e6abe5965c850379a7ce4807c431e4250be54e3881b83888bd"
				
				[versions]
				minecraft = "1.21.4"
				quilt = "0.28.0-beta.8"
			`),
			want: PackFile{
				Name:       "ExamplePack",
				Version:    "1.0.0",
				PackFormat: "packwiz:1.1.0",
				Index: PackFileIndex{
					File: "index.toml",
				},
				Versions: PackFileVersions{
					Minecraft:  "1.21.4",
					Quilt:      "0.28.0-beta.8",
					Forge:      "",
					Fabric:     "",
					LiteLoader: "",
					NeoForge:   "",
				},
			},
			wantErr: nil,
		},
		{
			name:    "empty data",
			data:    []byte(""),
			want:    PackFile{},
			wantErr: errors.New("pack.toml is invalid"),
		},
		{
			name:    "invalid toml format",
			data:    []byte("invalid_toml ="),
			want:    PackFile{},
			wantErr: errors.New("toml: expected value, not eof"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parsePackFile(tc.data)
			if tc.wantErr != nil {
				if err == nil || err.Error() != tc.wantErr.Error() {
					t.Errorf("expected error %v, got %v", tc.wantErr, err)
				}
			} else if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if err == nil && got != tc.want {
				t.Errorf("expected %v, got %v", tc.want, got)
			}
		})
	}
}
