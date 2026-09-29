import type {UpdateAllResponse} from "@/interfaces/pack.ts"

export type UpdateAllOutcome = "up-to-date" | "details"

// Fully up to date (nothing updated, skipped or failed) needs only a snackbar;
// anything else is shown in the result dialog.
export function updateAllOutcome(r: Pick<UpdateAllResponse, "updated" | "skipped" | "failed">): UpdateAllOutcome {
  return r.updated.length || r.skipped.length || r.failed.length ? "details" : "up-to-date"
}

function plural(n: number, word: string): string {
  return `${n} ${word}${n === 1 ? "" : "s"}`
}

// One-line summary, e.g. "2 mods updated, 1 failed, 1 skipped (pinned), 5 up to date".
export function summarizeUpdateAll(r: UpdateAllResponse): string {
  if (updateAllOutcome(r) === "up-to-date") return "All mods are up to date"
  const parts: string[] = []
  if (r.updated.length) parts.push(`${plural(r.updated.length, "mod")} updated`)
  if (r.failed.length) parts.push(`${r.failed.length} failed`)
  if (r.skipped.length) parts.push(`${r.skipped.length} skipped (pinned)`)
  if (r.upToDate) parts.push(`${r.upToDate} up to date`)
  return parts.join(", ")
}
