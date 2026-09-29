import {describe, expect, it} from "vitest"
import type {Mod} from "@/interfaces/pack.ts"
import {countMods, filterModsBySide, formatModCounts, matchesSide} from "./mod-filters.ts"

const mod = (side: Mod["side"], isDependency = false) => ({side, isDependency}) as Mod

describe("countMods", () => {
  it("counts empty list", () => {
    expect(countMods([])).toEqual({total: 0, dependencies: 0})
  })

  it("counts total and dependencies", () => {
    expect(countMods([mod("both"), mod("client", true), mod("server", true)]))
      .toEqual({total: 3, dependencies: 2})
  })
})

describe("side filtering", () => {
  it("keeps everything for empty/both filter", () => {
    const mods = [mod("client"), mod("server")]
    expect(filterModsBySide(mods, "")).toHaveLength(2)
    expect(matchesSide("client", "both")).toBe(true)
  })

  it("keeps matching and universal mods", () => {
    const mods = [mod("client"), mod("server"), mod("both")]
    expect(filterModsBySide(mods, "client")).toHaveLength(2)
  })
})

describe("formatModCounts", () => {
  it("handles zero", () => {
    expect(formatModCounts({total: 0, dependencies: 0})).toBe("0 mods · 0 dependencies")
  })

  it("handles singular", () => {
    expect(formatModCounts({total: 1, dependencies: 1})).toBe("1 mod · 1 dependency")
  })

  it("handles plural", () => {
    expect(formatModCounts({total: 5, dependencies: 2})).toBe("5 mods · 2 dependencies")
  })
})
