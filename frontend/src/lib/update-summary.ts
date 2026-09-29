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

const SKIP_REASON_LABELS: Record<string, string> = {
  pinned: "Pinned",
}

// Human label for a skip reason from the API; unknown reasons are capitalized.
export function skipReasonLabel(reason?: string): string {
  if (!reason) return ""
  return SKIP_REASON_LABELS[reason] ?? reason.charAt(0).toUpperCase() + reason.slice(1)
}

// One-line summary, e.g. "2 mods updated, 1 failed, 1 skipped, 5 up to date, 3 not checked (manual sources)".
export function summarizeUpdateAll(r: UpdateAllResponse): string {
  if (!r.notChecked && updateAllOutcome(r) === "up-to-date") return "All mods are up to date"
  const parts: string[] = []
  if (r.updated.length) parts.push(`${plural(r.updated.length, "mod")} updated`)
  if (r.failed.length) parts.push(`${r.failed.length} failed`)
  if (r.skipped.length) parts.push(`${r.skipped.length} skipped`)
  if (r.upToDate) parts.push(`${r.upToDate} up to date`)
  if (r.notChecked) parts.push(`${r.notChecked} not checked (manual sources)`)
  return parts.join(", ")
}
