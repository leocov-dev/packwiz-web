import type {BulkAddCandidate, BulkAddItem, BulkAddSource, BulkAddStatus, BulkItemStatus} from "@/interfaces/pack.ts"

/** Mirrors the server limit (packwiz_svc.maxBulkLines). */
export const BULK_MAX_LINES = 200

/** A bulk add can run up to the server job timeout (10 min) plus queueing. */
export const BULK_POLL_MAX_DURATION_MS = 12 * 60 * 1000

/**
 * Splits pasted text like the server does: trimmed, blank lines and "#"
 * comments dropped, repeats (ignoring case and spacing) dropped.
 */
export function parseBulkLines(text: string): string[] {
  const seen = new Set<string>()
  const out: string[] = []
  for (const raw of text.split("\n")) {
    const line = raw.trim().split(/\s+/).join(" ")
    if (!line || line.startsWith("#")) continue
    const key = line.toLowerCase()
    if (seen.has(key)) continue
    seen.add(key)
    out.push(line)
  }
  return out
}

export const isActiveBulkStatus = (status: BulkAddStatus | undefined): boolean =>
  status === "queued" || status === "running"

/** Lines waiting for the user to pick a mod or reject them. */
export const isReviewable = (status: BulkItemStatus): boolean =>
  status === "pending" || status === "incompatible_version" || status === "conflict"

export type BulkItemGroup = "all" | "review" | "added" | "skipped" | "failed"

export function itemGroup(status: BulkItemStatus): Exclude<BulkItemGroup, "all"> | "queued" {
  if (isReviewable(status)) return "review"
  switch (status) {
    case "added":
    case "picked":
      return "added"
    case "skipped":
    case "rejected":
      return "skipped"
    case "failed":
      return "failed"
    default:
      return "queued"
  }
}

export function groupCounts(items: BulkAddItem[]): Record<BulkItemGroup, number> {
  const counts: Record<BulkItemGroup, number> = {all: items.length, review: 0, added: 0, skipped: 0, failed: 0}
  for (const it of items) {
    const g = itemGroup(it.status)
    if (g !== "queued") counts[g]++
  }
  return counts
}

export function filterItems(items: BulkAddItem[], group: BulkItemGroup): BulkAddItem[] {
  return group === "all" ? items : items.filter(it => itemGroup(it.status) === group)
}

export interface BulkStatusMeta {
  label: string
  icon: string
  /** Semantic chip color; undefined for neutral. */
  color?: "success" | "warning" | "error"
}

const STATUS_META: Record<BulkItemStatus, BulkStatusMeta> = {
  queued: {label: "Queued", icon: "mdi-timer-sand"},
  added: {label: "Added", icon: "mdi-check-circle", color: "success"},
  picked: {label: "Added after review", icon: "mdi-check-circle", color: "success"},
  pending: {label: "Needs review", icon: "mdi-help-circle-outline", color: "warning"},
  incompatible_version: {label: "No compatible version", icon: "mdi-alert-outline", color: "warning"},
  conflict: {label: "Conflict", icon: "mdi-alert-octagon-outline", color: "warning"},
  skipped: {label: "Skipped", icon: "mdi-skip-next-circle-outline"},
  rejected: {label: "Rejected", icon: "mdi-close-circle-outline"},
  failed: {label: "Failed", icon: "mdi-alert-circle", color: "error"},
}

export function statusMeta(status: BulkItemStatus): BulkStatusMeta {
  return STATUS_META[status] ?? {label: status, icon: "mdi-help-circle-outline"}
}

export function sourceLabel(source: BulkAddSource | ""): string {
  if (source === "modrinth") return "Modrinth"
  if (source === "curseforge") return "CurseForge"
  return ""
}

/** Search to open the review of a line with: the source of its first stored candidate. */
export function initialReviewSource(item: Pick<BulkAddItem, "source" | "candidates">, curseforgeAvailable: boolean): BulkAddSource {
  const first: BulkAddCandidate | undefined = item.candidates?.[0]
  const source = item.source || first?.source || "modrinth"
  return source === "curseforge" && !curseforgeAvailable ? "modrinth" : source
}

/** Replaces one item (by id) in a run's item list, keeping order. */
export function replaceItem(items: BulkAddItem[], updated: BulkAddItem): BulkAddItem[] {
  return items.map(it => (it.id === updated.id ? updated : it))
}

export function progressText(status: BulkAddStatus, lineCount: number): string {
  const mods = `${lineCount} ${lineCount === 1 ? "mod" : "mods"}`
  return status === "queued" ? `Waiting to look up ${mods}...` : `Looking up ${mods}...`
}
