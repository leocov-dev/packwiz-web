package utils

import "testing"

func TestCompareMinecraftVersions(t *testing.T) {
	for _, tc := range []struct {
		a, b string
		want int
	}{
		{"1.21.1", "1.21.1", 0},
		{"1.21", "1.21.0", 0},
		{"1.20.4", "1.20.5", -1},
		{"1.21.10", "1.21.9", 1},
		{"1.21.1", "26.1", -1},
		{"26.1", "26.1.2", -1},
		{"26.3", "26.2", 1},
		{"26.3-rc-3", "26.3", -1},
		{"26.3-pre-1", "26.3-rc-1", -1},
		{"26.3-snapshot-10", "26.3-pre-1", -1},
		{"26.3-snapshot-10", "26.3-snapshot-9", 1},
		{"26.4-snapshot-1", "26.3", 1},
		{"1.21-pre1", "1.21", -1},
		{"1.21.5-rc1", "1.21.5-pre2", 1},
	} {
		got, ok := CompareMinecraftVersions(tc.a, tc.b)
		if !ok || got != tc.want {
			t.Errorf("Compare(%q, %q) = %d, %v; want %d", tc.a, tc.b, got, ok, tc.want)
		}
	}
}

func TestCompareMinecraftVersionsUnparseable(t *testing.T) {
	for _, pair := range [][2]string{
		{"24w14a", "1.21"},
		{"1.21", ""},
		{"26", "26.1"},
		{"26.1-beta-1", "26.1"},
		{"a.b", "1.2"},
	} {
		if _, ok := CompareMinecraftVersions(pair[0], pair[1]); ok {
			t.Errorf("Compare(%q, %q) should not be comparable", pair[0], pair[1])
		}
	}
}
