import type {Mod} from "@/interfaces/pack.ts"

export type ModSide = "client" | "server" | "both" | ""

// Mirrors packwiz-nxt/cmd/list.go's side-filter fallthrough: a mod is kept
// if its side matches the requested side, or either side is universal
// ("both"/""), since those mean "applies regardless of side".
export function matchesSide(modSide: ModSide, filterSide: ModSide): boolean {
  if (filterSide === "" || filterSide === "both") return true
  return modSide === filterSide || modSide === "" || modSide === "both"
}

export function filterModsBySide(mods: Mod[], filterSide: ModSide): Mod[] {
  if (!filterSide) return mods
  return mods.filter(mod => matchesSide(mod.side, filterSide))
}

export interface ModCounts {
  total: number
  dependencies: number
}

export function countMods(mods: Mod[]): ModCounts {
  return {
    total: mods.length,
    dependencies: mods.filter(mod => mod.isDependency).length,
  }
}

const plural = (n: number, singular: string, pluralForm: string) => `${n} ${n === 1 ? singular : pluralForm}`

export function formatModCounts(counts: ModCounts): string {
  return `${plural(counts.total, "mod", "mods")} · ${plural(counts.dependencies, "dependency", "dependencies")}`
}

// ---- source / side labels ----

// Values of `mod.source` as set by the backend (tables.ExtractModSource).
export function modSourceLabel(source: string | undefined): string {
  switch ((source ?? "").toLowerCase()) {
    case "modrinth": return "Modrinth"
    case "curseforge": return "CurseForge"
    case "github": return "GitHub"
    default: return ""
  }
}

export function modSideLabel(side: string | undefined): string {
  switch (side) {
    case "client": return "Client"
    case "server": return "Server"
    case "both": return "Client + Server"
    default: return ""
  }
}

// ---- dependency relationships ----
// `mod.dependencyIds` lists the mods THIS mod depends on (set on the
// requiring mod at add time), so "required by" is the inverse lookup.

export type DependentsMap = Map<number, Mod[]>

export function buildDependentsMap(mods: Mod[]): DependentsMap {
  const map: DependentsMap = new Map()
  for (const mod of mods) {
    for (const depId of mod.dependencyIds ?? []) {
      const list = map.get(depId)
      if (list) list.push(mod)
      else map.set(depId, [mod])
    }
  }
  return map
}

export function dependentNames(mod: Mod, dependents: DependentsMap): string[] {
  return (dependents.get(mod.id) ?? []).map(m => m.name).sort((a, b) => a.localeCompare(b))
}

export function dependencyTooltip(names: string[]): string {
  if (names.length === 0) return "Installed automatically as a dependency"
  return `Required by ${names.join(", ")}`
}

// Automatically-installed dependencies of `mod` that no other mod requires.
// The backend does not delete them when `mod` is removed.
export function findOrphanedDependencies(mod: Mod, mods: Mod[], dependents: DependentsMap): Mod[] {
  const byId = new Map(mods.map(m => [m.id, m]))
  const orphans: Mod[] = []
  for (const depId of mod.dependencyIds ?? []) {
    const dep = byId.get(depId)
    if (!dep || !dep.isDependency) continue
    const requirers = dependents.get(depId) ?? []
    if (requirers.every(r => r.id === mod.id)) orphans.push(dep)
  }
  return orphans.sort((a, b) => a.name.localeCompare(b.name))
}

export function removeModMessage(modName: string, orphanNames: string[]): string {
  const base = `Are you sure you want to remove ${modName}?`
  if (orphanNames.length === 0) return base
  const label = orphanNames.length === 1 ? "dependency" : "dependencies"
  return `${base} This will also leave ${orphanNames.length} ${label} unused: ${orphanNames.join(", ")}. They stay installed until removed separately.`
}

// ---- search / show-filters / sorting ----

export type ModShow = "pinned" | "optional" | "dependencies"
export type ModSort = "name" | "updated"

export const MOD_SHOW_VALUES: ModShow[] = ["pinned", "optional", "dependencies"]
export const MOD_SORT_VALUES: ModSort[] = ["name", "updated"]

export function searchMods(mods: Mod[], query: string): Mod[] {
  const q = (query ?? "").trim().toLowerCase()
  if (!q) return mods
  return mods.filter(mod =>
    [mod.name, mod.slug, mod.fileName].some(field => (field ?? "").toLowerCase().includes(q)))
}

function matchesShow(mod: Mod, show: ModShow): boolean {
  switch (show) {
    case "pinned": return !!mod.pinned
    case "optional": return !!mod.option?.optional
    case "dependencies": return !!mod.isDependency
  }
}

// Union semantics: a mod is kept if it matches ANY selected category.
// No selection keeps everything.
export function filterModsByShow(mods: Mod[], show: ModShow[]): Mod[] {
  if (show.length === 0) return mods
  return mods.filter(mod => show.some(s => matchesShow(mod, s)))
}

const timeOf = (iso: string | undefined) => {
  const t = Date.parse(iso ?? "")
  return Number.isNaN(t) ? 0 : t
}

export function sortMods(mods: Mod[], sort: ModSort): Mod[] {
  const cmp = sort === "updated"
    ? (a: Mod, b: Mod) => timeOf(b.updatedAt) - timeOf(a.updatedAt) || a.name.localeCompare(b.name)
    : (a: Mod, b: Mod) => a.name.localeCompare(b.name)
  const regular = mods.filter(m => !m.isDependency).sort(cmp)
  const deps = mods.filter(m => m.isDependency).sort(cmp)
  return [...regular, ...deps]
}

export interface ModListState {
  q: string
  sort: ModSort
  side: ModSide
  show: ModShow[]
}

export const DEFAULT_MOD_LIST_STATE: ModListState = {q: "", sort: "name", side: "", show: []}

export function applyModListState(mods: Mod[], state: ModListState): Mod[] {
  return sortMods(
    filterModsByShow(filterModsBySide(searchMods(mods, state.q), state.side), state.show),
    state.sort,
  )
}

export function hasActiveModFilters(state: ModListState): boolean {
  return !!state.q.trim() || !!state.side || state.show.length > 0
}

const first = (v: unknown): string => {
  const value = Array.isArray(v) ? v[0] : v
  return typeof value === "string" ? value : ""
}

export function parseModListQuery(query: Record<string, unknown>): ModListState {
  const sort = first(query.sort) as ModSort
  const side = first(query.side) as ModSide
  const show = first(query.show).split(",").filter((s): s is ModShow => MOD_SHOW_VALUES.includes(s as ModShow))
  return {
    q: first(query.q),
    sort: MOD_SORT_VALUES.includes(sort) ? sort : DEFAULT_MOD_LIST_STATE.sort,
    side: ["client", "server", "both"].includes(side) ? side : "",
    show: [...new Set(show)],
  }
}

// Only non-default values are emitted so the URL stays clean.
export function buildModListQuery(state: ModListState): Record<string, string> {
  const out: Record<string, string> = {}
  if (state.q) out.q = state.q
  if (state.sort !== DEFAULT_MOD_LIST_STATE.sort) out.sort = state.sort
  if (state.side) out.side = state.side
  if (state.show.length > 0) out.show = state.show.join(",")
  return out
}

export function formatFilteredCount(shown: number, total: number): string {
  return `${shown} of ${plural(total, "mod", "mods")}`
}
