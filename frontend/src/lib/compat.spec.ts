import {describe, expect, it} from "vitest"
import type {CompatFindingItem, CompatStatus} from "@/interfaces/compat.ts"
import {
  canRescan,
  compatCountsText,
  compatDisplay,
  compatUnscannedText,
  findingTitle,
  formatScannedAgo,
  groupFindings,
  isCompatActive,
  shouldContinueCompatPolling,
  COMPAT_POLL_MAX_DURATION_MS,
} from "@/lib/compat.ts"

const f = (over: Partial<CompatFindingItem>): CompatFindingItem => ({
  modId: 1, slug: "a", severity: "error", code: "wrong_loader", message: "m", detail: {}, ...over,
})

describe("compatDisplay", () => {
  const statuses: CompatStatus[] = ["unknown", "queued", "scanning", "stale", "ok", "warnings", "errors", "failed"]
  it.each(statuses)("maps %s to an icon, color and label", (s) => {
    const d = compatDisplay(s, {errors: 2, warnings: 1})
    expect(d.icon).toMatch(/^mdi-/)
    expect(d.color).not.toBe("")
    expect(d.label).not.toBe("")
  })
  it("uses semantic colors", () => {
    expect(compatDisplay("ok").color).toBe("success")
    expect(compatDisplay("warnings").color).toBe("warning")
    expect(compatDisplay("errors").color).toBe("error")
    expect(compatDisplay("stale").color).toBe("grey")
  })
  it("spins only while queued or scanning", () => {
    expect(statuses.filter(s => compatDisplay(s).busy)).toEqual(["queued", "scanning"])
  })
  it("includes counts in the label", () => {
    expect(compatDisplay("errors", {errors: 2, warnings: 1}).label).toContain("2 errors, 1 warning")
  })
  it("failed is outlined", () => {
    expect(compatDisplay("failed").variant).toBe("outlined")
  })
})

describe("compatCountsText", () => {
  it("pluralizes and skips zeros", () => {
    expect(compatCountsText({errors: 1, warnings: 0})).toBe("1 error")
    expect(compatCountsText({errors: 2, warnings: 1})).toBe("2 errors, 1 warning")
    expect(compatCountsText({errors: 0, warnings: 0})).toBe("")
    expect(compatCountsText(undefined)).toBe("")
  })
})

describe("findingTitle", () => {
  it("maps known codes and falls back to the code", () => {
    expect(findingTitle("missing_dependency")).toBe("Missing dependency")
    expect(findingTitle("brand_new_code")).toBe("brand_new_code")
  })
})

describe("groupFindings", () => {
  it("groups by severity then mod, in severity order", () => {
    const groups = groupFindings([
      f({severity: "info", modId: 3, slug: "c"}),
      f({severity: "warning", modId: 2, slug: "b"}),
      f({severity: "error", modId: 1, slug: "a", code: "x"}),
      f({severity: "error", modId: 1, slug: "a", code: "y"}),
      f({severity: "error", modId: 2, slug: "b"}),
    ])
    expect(groups.map(g => g.severity)).toEqual(["error", "warning", "info"])
    expect(groups[0].total).toBe(3)
    expect(groups[0].mods.map(m => m.slug)).toEqual(["a", "b"])
    expect(groups[0].mods[0].findings.map(x => x.code)).toEqual(["x", "y"])
  })
  it("groups findings without mod id by slug and drops empty severities", () => {
    const groups = groupFindings([f({modId: null, slug: "x"}), f({modId: null, slug: "x"}), f({modId: null, slug: "y"})])
    expect(groups).toHaveLength(1)
    expect(groups[0].mods.map(m => m.slug)).toEqual(["x", "y"])
  })
  it("returns nothing for no findings", () => {
    expect(groupFindings([])).toEqual([])
  })
})

describe("misc helpers", () => {
  it("unscanned text", () => {
    expect(compatUnscannedText(0)).toBe("")
    expect(compatUnscannedText(1)).toBe("1 mod could not be checked")
    expect(compatUnscannedText(3)).toBe("3 mods could not be checked")
  })
  it("scanned ago", () => {
    const now = Date.parse("2026-01-02T00:00:00Z")
    expect(formatScannedAgo(null, now)).toBe("Never scanned")
    expect(formatScannedAgo("2026-01-02T00:00:00Z", now)).toBe("Last scanned just now")
    expect(formatScannedAgo("2026-01-01T23:30:00Z", now)).toBe("Last scanned 30 min ago")
    expect(formatScannedAgo("2026-01-01T21:00:00Z", now)).toBe("Last scanned 3 h ago")
    expect(formatScannedAgo("2025-12-30T00:00:00Z", now)).toBe("Last scanned 3 d ago")
  })
  it("polls only while active and under the cap", () => {
    expect(isCompatActive("queued")).toBe(true)
    expect(isCompatActive("ok")).toBe(false)
    expect(shouldContinueCompatPolling("scanning", 1000)).toBe(true)
    expect(shouldContinueCompatPolling("scanning", COMPAT_POLL_MAX_DURATION_MS)).toBe(false)
    expect(shouldContinueCompatPolling("ok", 0)).toBe(false)
  })
  it("rescan gating", () => {
    expect(canRescan(true, false)).toBe(true)
    expect(canRescan(false, false)).toBe(false)
    expect(canRescan(true, true)).toBe(false)
  })
})
