import type {Mod} from "@/interfaces/pack.ts"
import type {DependentsMap} from "@/lib/mod-filters.ts"

export interface RemovalPlan {
  /** Ids to send for removal, selection first. */
  ids: number[]
  /** Dependencies that no remaining mod requires once the selection is removed. */
  unused: Mod[]
  /** Selected mods still required by a mod that stays, as "Name (required by A, B)". */
  blocked: string[]
}

const byName = (a: Mod, b: Mod) => a.name.localeCompare(b.name)

/**
 * Plans removing the selected mods. Dependencies left unused (also through a
 * chain of dependencies) are reported, and added to the removal when
 * includeUnused. A selected mod that a staying mod still requires blocks the
 * removal, as the server refuses it.
 */
export function planRemoval(mods: Mod[], selected: ReadonlySet<number>, dependents: DependentsMap, includeUnused: boolean): RemovalPlan {
  const removing = new Set([...selected].filter(id => mods.some(m => m.id === id)))

  const unused: Mod[] = []
  let changed = true
  while (changed) {
    changed = false
    for (const mod of mods) {
      if (!mod.isDependency || removing.has(mod.id) || unused.includes(mod)) continue
      const requirers = dependents.get(mod.id) ?? []
      const gone = (r: Mod) => removing.has(r.id) || unused.includes(r)
      if (requirers.length > 0 && requirers.every(gone)) {
        unused.push(mod)
        changed = true
      }
    }
  }

  const final = new Set(removing)
  if (includeUnused) for (const m of unused) final.add(m.id)

  const blocked: string[] = []
  for (const mod of mods) {
    if (!final.has(mod.id)) continue
    const staying = [...new Set((dependents.get(mod.id) ?? []).filter(r => !final.has(r.id)).map(r => r.name))]
    if (staying.length > 0) {
      blocked.push(`${mod.name} (required by ${staying.sort((a, b) => a.localeCompare(b)).join(", ")})`)
    }
  }

  return {ids: [...final], unused: unused.sort(byName), blocked: blocked.sort((a, b) => a.localeCompare(b))}
}

export function removeModsMessage(count: number, unusedCount: number, includeUnused: boolean): string {
  const mods = `${count} ${count === 1 ? "mod" : "mods"}`
  if (unusedCount === 0) return `Remove ${mods} from the pack?`
  const deps = `${unusedCount} ${unusedCount === 1 ? "dependency" : "dependencies"}`
  return includeUnused
    ? `Remove ${mods} and ${deps} no longer needed?`
    : `Remove ${mods}? ${deps} will stay installed but unused.`
}

/** Drops selected ids that are no longer in the list (e.g. after a reload). */
export function pruneSelection(selected: ReadonlySet<number>, mods: Mod[]): Set<number> {
  const ids = new Set(mods.map(m => m.id))
  return new Set([...selected].filter(id => ids.has(id)))
}

export function toggleSelected(selected: ReadonlySet<number>, id: number): Set<number> {
  const next = new Set(selected)
  if (next.has(id)) next.delete(id)
  else next.add(id)
  return next
}

/** Selects every given mod, or clears them all when all are already selected. */
export function toggleAll(selected: ReadonlySet<number>, mods: Mod[]): Set<number> {
  const next = new Set(selected)
  const allSelected = mods.length > 0 && mods.every(m => next.has(m.id))
  for (const m of mods) {
    if (allSelected) next.delete(m.id)
    else next.add(m.id)
  }
  return next
}
