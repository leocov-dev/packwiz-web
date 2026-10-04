package utils

import (
	"strconv"
	"strings"
)

// mcStage orders the kinds of Minecraft builds of one version.
var mcStage = map[string]int{
	"snapshot": 0,
	"pre":      1,
	"rc":       2,
}

const mcReleaseStage = 3

type mcVersion struct {
	nums  []int
	stage int
	build int
}

// parseMinecraftVersion reads a release ("1.21.1", "26.1", "26.1.2") or a
// development build of one ("26.4-snapshot-2", "26.3-pre-1", "26.3-rc-3",
// "1.21-pre1", "1.21.5-rc1"). Weekly snapshot ids like "24w14a" and anything
// else are not parseable.
func parseMinecraftVersion(v string) (mcVersion, bool) {
	core, suffix, hasSuffix := strings.Cut(strings.TrimSpace(v), "-")

	parts := strings.Split(core, ".")
	if len(parts) < 2 {
		return mcVersion{}, false
	}
	out := mcVersion{nums: make([]int, 0, len(parts)), stage: mcReleaseStage}
	for _, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil || n < 0 {
			return mcVersion{}, false
		}
		out.nums = append(out.nums, n)
	}

	if !hasSuffix {
		return out, true
	}

	// "snapshot-2", "pre-1", "rc-3" (year-based) or "pre1", "rc1" (older)
	suffix = strings.ReplaceAll(suffix, "-", "")
	kind := strings.TrimRight(suffix, "0123456789")
	stage, ok := mcStage[kind]
	if !ok {
		return mcVersion{}, false
	}
	build := 0
	if digits := suffix[len(kind):]; digits != "" {
		n, err := strconv.Atoi(digits)
		if err != nil {
			return mcVersion{}, false
		}
		build = n
	}
	out.stage, out.build = stage, build
	return out, true
}

// CompareMinecraftVersions orders two Minecraft version ids: -1 if a is older
// than b, 0 if equal, 1 if newer. Year-based ids ("26.1") sort after all
// "1.x" ids because the numbers compare naturally. A development build sorts
// before its release (snapshot < pre < rc < release). ok is false when either
// id cannot be parsed (e.g. a weekly snapshot like "24w14a"); the result is
// then meaningless.
func CompareMinecraftVersions(a, b string) (cmp int, ok bool) {
	va, okA := parseMinecraftVersion(a)
	vb, okB := parseMinecraftVersion(b)
	if !okA || !okB {
		return 0, false
	}

	for i := 0; i < max(len(va.nums), len(vb.nums)); i++ {
		x, y := 0, 0
		if i < len(va.nums) {
			x = va.nums[i]
		}
		if i < len(vb.nums) {
			y = vb.nums[i]
		}
		if x != y {
			return sign(x - y), true
		}
	}
	if va.stage != vb.stage {
		return sign(va.stage - vb.stage), true
	}
	return sign(va.build - vb.build), true
}

func sign(n int) int {
	switch {
	case n < 0:
		return -1
	case n > 0:
		return 1
	}
	return 0
}
