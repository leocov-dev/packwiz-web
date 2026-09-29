import {describe, expect, it} from "vitest"
import type {Mod} from "@/interfaces/pack.ts"
import {countMods, filterModsBySide, matchesSide} from "./mod-filters.ts"

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
