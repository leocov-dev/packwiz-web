import {describe, expect, it} from "vitest"
import type {Mod} from "@/interfaces/pack.ts"
import {buildDependentsMap} from "@/lib/mod-filters.ts"
import {planRemoval, pruneSelection, removeModsMessage, toggleAll, toggleSelected} from "@/lib/mod-selection.ts"

const m = (id: number, name: string, dependencyIds: number[] = [], isDependency = false) =>
  ({id, name, dependencyIds, isDependency}) as Mod

// Lithium -> Fabric API; Sodium -> Fabric API, Indium; Indium -> Core
const mods = [
  m(1, "Lithium", [3]),
  m(2, "Sodium", [3, 4]),
  m(3, "Fabric API", [], true),
  m(4, "Indium", [5], true),
  m(5, "Core", [], true),
  m(6, "JEI"),
]
const deps = buildDependentsMap(mods)

describe("planRemoval", () => {
  it("removes plain mods", () => {
    const p = planRemoval(mods, new Set([6]), deps, true)
    expect(p).toEqual({ids: [6], unused: [], blocked: []})
  })

  it("finds dependencies left unused, through chains", () => {
    const p = planRemoval(mods, new Set([2]), deps, true)
    expect(p.unused.map(x => x.name)).toEqual(["Core", "Indium"])
    expect(p.ids.sort()).toEqual([2, 4, 5])
    expect(p.blocked).toEqual([])
  })

  it("keeps unused dependencies when not included", () => {
    const p = planRemoval(mods, new Set([2]), deps, false)
    expect(p.ids).toEqual([2])
    expect(p.unused).toHaveLength(2)
  })

  it("a dependency still needed by another mod is not unused", () => {
    const p = planRemoval(mods, new Set([1]), deps, true)
    expect(p.unused).toEqual([])
  })

  it("blocks removing a dependency still required", () => {
    const p = planRemoval(mods, new Set([1, 3]), deps, true)
    expect(p.blocked).toEqual(["Fabric API (required by Sodium)"])
  })

  it("allows a dependency together with all its requirers", () => {
    const p = planRemoval(mods, new Set([1, 2, 3]), deps, true)
    expect(p.blocked).toEqual([])
    expect(p.ids.sort()).toEqual([1, 2, 3, 4, 5])
  })

  it("ignores ids not in the list", () => {
    expect(planRemoval(mods, new Set([99]), deps, true).ids).toEqual([])
  })
})

describe("removeModsMessage", () => {
  it("phrases counts", () => {
    expect(removeModsMessage(1, 0, true)).toBe("Remove 1 mod from the pack?")
    expect(removeModsMessage(3, 2, true)).toBe("Remove 3 mods and 2 dependencies no longer needed?")
    expect(removeModsMessage(2, 1, false)).toBe("Remove 2 mods? 1 dependency will stay installed but unused.")
  })
})

describe("selection helpers", () => {
  it("toggles one", () => {
    expect([...toggleSelected(new Set([1]), 2)]).toEqual([1, 2])
    expect([...toggleSelected(new Set([1, 2]), 1)]).toEqual([2])
  })
  it("toggles all shown", () => {
    const shown = [mods[0], mods[1]]
    expect([...toggleAll(new Set([6]), shown)].sort()).toEqual([1, 2, 6])
    expect([...toggleAll(new Set([1, 2, 6]), shown)]).toEqual([6])
  })
  it("prunes ids that are gone", () => {
    expect([...pruneSelection(new Set([1, 99]), mods)]).toEqual([1])
  })
})
