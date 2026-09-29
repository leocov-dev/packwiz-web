import {describe, expect, it} from "vitest"
import type {Mod} from "@/interfaces/pack.ts"
import {
  applyModListState, applyPinOverrides, buildDependentNamesMap, buildOrphanNamesMap, buildDependentsMap, buildModListQuery, countMods, DEFAULT_MOD_LIST_STATE, dependencyTooltip,
  dependentNames, filterModsBySide, filterModsByShow, findOrphanedDependencies, formatFilteredCount, formatModCounts,
  displayVersion, hasActiveModFilters, modVersion, matchesSide, modSideLabel, modSourceLabel, parseModListQuery, removeModMessage, searchMods,
  sortMods,
} from "./mod-filters.ts"

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

const full = (o: Partial<Mod> & {id: number}) => ({
  name: `mod${o.id}`, slug: `slug${o.id}`, fileName: `file${o.id}.jar`, side: "both",
  isDependency: false, pinned: false, updatedAt: "2024-01-01T00:00:00Z", ...o,
}) as Mod

describe("source and side labels", () => {
  it("maps backend source values", () => {
    expect(modSourceLabel("modrinth")).toBe("Modrinth")
    expect(modSourceLabel("curseforge")).toBe("CurseForge")
    expect(modSourceLabel("github")).toBe("GitHub")
    expect(modSourceLabel("")).toBe("")
    expect(modSourceLabel("other")).toBe("")
  })
  it("labels sides", () => {
    expect(modSideLabel("client")).toBe("Client")
    expect(modSideLabel("server")).toBe("Server")
    expect(modSideLabel("both")).toBe("Client + Server")
    expect(modSideLabel("x")).toBe("")
  })
})

describe("dependents", () => {
  const dep = full({id: 1, name: "Lib", isDependency: true})
  const shared = full({id: 2, name: "Shared", isDependency: true})
  const a = full({id: 3, name: "Alpha", dependencyIds: [1, 2]})
  const b = full({id: 4, name: "Beta", dependencyIds: [2]})
  const mods = [dep, shared, a, b]
  const map = buildDependentsMap(mods)

  it("builds inverse map", () => {
    expect(dependentNames(shared, map)).toEqual(["Alpha", "Beta"])
    expect(dependentNames(dep, map)).toEqual(["Alpha"])
    expect(dependentNames(a, map)).toEqual([])
  })
  it("formats tooltip", () => {
    expect(dependencyTooltip(["A", "B"])).toBe("Required by A, B")
    expect(dependencyTooltip([])).toBe("No mod requires this dependency anymore")
  })
  it("finds orphaned dependencies only", () => {
    expect(findOrphanedDependencies(a, mods, map).map(m => m.name)).toEqual(["Lib"])
    expect(findOrphanedDependencies(b, mods, map)).toEqual([])
  })
  it("ignores non-dependency and missing ids", () => {
    const regular = full({id: 5, name: "Reg"})
    const c = full({id: 6, dependencyIds: [5, 99]})
    const ms = [regular, c]
    expect(findOrphanedDependencies(c, ms, buildDependentsMap(ms))).toEqual([])
  })
  it("builds remove message", () => {
    expect(removeModMessage("X", [])).toBe("Are you sure you want to remove X?")
    expect(removeModMessage("X", ["Lib"])).toContain("also leave 1 dependency unused: Lib")
    expect(removeModMessage("X", ["A", "B"])).toContain("2 dependencies unused: A, B")
    expect(removeModMessage("Lib", [], true)).toBe("Remove unused dependency Lib? No mod requires this dependency anymore.")
  })
  it("ignores self-references and dedupes duplicate ids", () => {
    const lib = full({id: 1, name: "Lib", isDependency: true, dependencyIds: [1]})
    const x = full({id: 2, name: "X", dependencyIds: [1, 1]})
    const ms = [lib, x]
    const m = buildDependentsMap(ms)
    expect(dependentNames(lib, m)).toEqual(["X"])
    expect(m.get(1)).toHaveLength(1)
    expect(findOrphanedDependencies(x, ms, m).map(o => o.name)).toEqual(["Lib"])
    expect(findOrphanedDependencies(lib, ms, m)).toEqual([])
    expect(buildDependentNamesMap(buildDependentsMap([lib, x, full({id: 4, name: "X", dependencyIds: [1]})])).get(1)).toEqual(["X"])
  })
  it("does not orphan a dep also required by a regular mod", () => {
    const lib = full({id: 1, name: "Lib", isDependency: true})
    const x = full({id: 2, name: "X", dependencyIds: [1]})
    const y = full({id: 3, name: "Y", dependencyIds: [1]})
    const ms = [lib, x, y]
    expect(findOrphanedDependencies(x, ms, buildDependentsMap(ms))).toEqual([])
  })
  it("orphans a dep whose only other requirer is being removed via another dep chain", () => {
    const lib = full({id: 1, name: "Lib", isDependency: true})
    const mid = full({id: 2, name: "Mid", isDependency: true, dependencyIds: [1]})
    const x = full({id: 3, name: "X", dependencyIds: [1, 2]})
    const ms = [lib, mid, x]
    // Mid still requires Lib, so removing X strands Mid only.
    expect(findOrphanedDependencies(x, ms, buildDependentsMap(ms)).map(m => m.name)).toEqual(["Mid"])
  })
  it("builds per-mod name maps once", () => {
    const lib = full({id: 1, name: "Lib", isDependency: true})
    const x = full({id: 2, name: "X", dependencyIds: [1]})
    const ms = [lib, x]
    const d = buildDependentsMap(ms)
    expect(buildDependentNamesMap(d).get(1)).toEqual(["X"])
    expect(buildDependentNamesMap(d).get(2)).toBeUndefined()
    expect(buildOrphanNamesMap(ms, d).get(2)).toEqual(["Lib"])
    expect(buildOrphanNamesMap([], new Map()).size).toBe(0)
  })
  it("handles empty mods", () => {
    expect(buildDependentsMap([]).size).toBe(0)
    expect(applyModListState([], DEFAULT_MOD_LIST_STATE)).toEqual([])
  })
  it("applies pin overrides without mutating", () => {
    const a = full({id: 1, pinned: false})
    const b = full({id: 2, pinned: true})
    const ms = [a, b]
    expect(applyPinOverrides(ms, new Map())).toBe(ms)
    const out = applyPinOverrides(ms, new Map([[1, true], [2, true], [99, true]]))
    expect(out.map(m => m.pinned)).toEqual([true, true])
    expect(out[1]).toBe(b)
    expect(a.pinned).toBe(false)
    const pinnedOnly = applyModListState(out, {...DEFAULT_MOD_LIST_STATE, show: ["pinned"]})
    expect(pinnedOnly.map(m => m.id)).toEqual([1, 2])
  })
})

describe("search, show filters and sort", () => {
  const mods = [
    full({id: 1, name: "Sodium", slug: "sodium", fileName: "sodium-1.jar", pinned: true, updatedAt: "2024-06-01T00:00:00Z"}),
    full({id: 2, name: "Iris", slug: "iris-shaders", fileName: "iris.jar", option: {optional: true, description: "", default: true}, updatedAt: "2024-05-01T00:00:00Z"}),
    full({id: 3, name: "Lib", slug: "lib", fileName: "lib.jar", isDependency: true, updatedAt: "2024-09-01T00:00:00Z"}),
  ]
  it("searches name, slug and file name", () => {
    expect(searchMods(mods, "SOD").map(m => m.id)).toEqual([1])
    expect(searchMods(mods, "shaders").map(m => m.id)).toEqual([2])
    expect(searchMods(mods, "lib.jar").map(m => m.id)).toEqual([3])
    expect(searchMods(mods, "  ")).toHaveLength(3)
  })
  it("searches version", () => {
    const withVersion = [{id: 1, name: "A", slug: "a", fileName: "a.jar", version: "1.20.4-0.5.8"}, {id: 2, name: "B", slug: "b", fileName: "b.jar"}] as Mod[]
    expect(searchMods(withVersion, "0.5.8").map(m => m.id)).toEqual([1])
  })
  it("filters by updates using the provided ids, unioned with other categories", () => {
    expect(filterModsByShow(mods, ["updates"], new Set([2])).map(m => m.id)).toEqual([2])
    expect(filterModsByShow(mods, ["updates"]).map(m => m.id)).toEqual([])
    expect(filterModsByShow(mods, ["pinned", "updates"], new Set([3])).map(m => m.id)).toEqual([1, 3])
  })
  it("persists updates in the route query", () => {
    const state = {...DEFAULT_MOD_LIST_STATE, show: ["updates" as const]}
    expect(buildModListQuery(state)).toEqual({show: "updates"})
    expect(parseModListQuery({show: "updates"}).show).toEqual(["updates"])
  })
  it("filters by show with union semantics", () => {
    expect(filterModsByShow(mods, [])).toHaveLength(3)
    expect(filterModsByShow(mods, ["pinned"]).map(m => m.id)).toEqual([1])
    expect(filterModsByShow(mods, ["optional", "dependencies"]).map(m => m.id)).toEqual([2, 3])
  })
  it("sorts by name and keeps dependencies last", () => {
    expect(sortMods(mods, "name").map(m => m.id)).toEqual([2, 1, 3])
  })
  it("sorts by recently updated and keeps dependencies last", () => {
    // Sodium (newest) sorts before Iris here, unlike name order (Iris, Sodium).
    expect(sortMods(mods, "updated").map(m => m.id)).toEqual([1, 2, 3])
    expect(sortMods(mods, "name").map(m => m.id)).toEqual([2, 1, 3])
    const swapped = [mods[0], {...mods[1], updatedAt: "2025-01-01T00:00:00Z"} as Mod, mods[2]]
    expect(sortMods(swapped, "updated").map(m => m.id)).toEqual([2, 1, 3])
  })
  it("treats invalid updatedAt as oldest", () => {
    const ms = [
      full({id: 1, name: "A", updatedAt: "garbage"}),
      full({id: 2, name: "B", updatedAt: "2024-01-01T00:00:00Z"}),
      full({id: 3, name: "C", updatedAt: undefined as unknown as string}),
    ]
    expect(sortMods(ms, "updated").map(m => m.id)).toEqual([2, 1, 3])
  })
  it("copes with garbage q and empty strings", () => {
    expect(searchMods(mods, undefined as unknown as string)).toHaveLength(3)
    expect(parseModListQuery({q: [1], sort: "", show: ""}))
      .toEqual({q: "", sort: "name", side: "", show: []})
  })
  it("applies whole state and reports activity", () => {
    const state = {...DEFAULT_MOD_LIST_STATE, show: ["pinned" as const]}
    expect(applyModListState(mods, state).map(m => m.id)).toEqual([1])
    expect(hasActiveModFilters(state)).toBe(true)
    expect(hasActiveModFilters(DEFAULT_MOD_LIST_STATE)).toBe(false)
    expect(hasActiveModFilters({...DEFAULT_MOD_LIST_STATE, side: "both"})).toBe(false)
    expect(hasActiveModFilters({...DEFAULT_MOD_LIST_STATE, side: "client"})).toBe(true)
  })
})

describe("list query persistence", () => {
  it("omits defaults", () => {
    expect(buildModListQuery(DEFAULT_MOD_LIST_STATE)).toEqual({})
  })
  it("round-trips", () => {
    const state = {q: "iris", sort: "updated" as const, side: "client" as const, show: ["pinned" as const, "optional" as const]}
    const query = buildModListQuery(state)
    expect(query).toEqual({q: "iris", sort: "updated", side: "client", show: "pinned,optional"})
    expect(parseModListQuery(query)).toEqual(state)
  })
  it("ignores invalid values and takes first of arrays", () => {
    expect(parseModListQuery({sort: "zzz", side: "x", show: "pinned,bogus,pinned", q: ["a", "b"]}))
      .toEqual({q: "a", sort: "name", side: "", show: ["pinned"]})
  })
  it("merges repeated show params", () => {
    expect(parseModListQuery({show: ["pinned", "optional,dependencies", "bogus"]}).show)
      .toEqual(["pinned", "optional", "dependencies"])
  })
  it("formats filtered count", () => {
    expect(formatFilteredCount(2, 10)).toBe("2 of 10 mods")
  })
})

describe("displayVersion", () => {
  it("prefers the stored version", () => {
    expect(displayVersion({version: " 1.2.3 ", fileName: "a-1.0.jar"})).toBe("1.2.3")
    expect(modVersion({version: undefined})).toBe("")
  })
  it("falls back to file name without .jar/.zip", () => {
    expect(displayVersion({fileName: "sodium-0.5.jar"})).toBe("sodium-0.5")
    expect(displayVersion({version: "", fileName: "pack.ZIP"})).toBe("pack")
    expect(displayVersion({fileName: "notes.txt"})).toBe("notes.txt")
  })
  it("is empty when nothing is known", () => {
    expect(displayVersion({fileName: ""})).toBe("")
    expect(displayVersion({} as Mod)).toBe("")
  })
})
