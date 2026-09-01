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
