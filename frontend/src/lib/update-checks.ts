import type {Mod, UpdateCheckItem, UpdateCheckStatus} from "@/interfaces/pack.ts"

export type UpdateChecksMap = ReadonlyMap<number, UpdateCheckItem>

export const POLL_BASE_MS = 2000
export const POLL_MAX_DELAY_MS = 8000
export const POLL_MAX_DURATION_MS = 5 * 60 * 1000

export const isActiveStatus = (status: UpdateCheckStatus): boolean =>
  status === "queued" || status === "running"

// ~2s at first, easing up to POLL_MAX_DELAY_MS.
export function nextPollDelay(attempt: number): number {
  return Math.min(POLL_BASE_MS * Math.pow(1.25, Math.max(0, attempt)), POLL_MAX_DELAY_MS)
}

export function shouldContinuePolling(status: UpdateCheckStatus, elapsedMs: number): boolean {
  return isActiveStatus(status) && elapsedMs < POLL_MAX_DURATION_MS
}

export function buildResultsMap(results: UpdateCheckItem[] | undefined): Map<number, UpdateCheckItem> {
  return new Map((results ?? []).map(r => [r.modId, r]))
}

// Mods with a known update that Update All would actually touch (not pinned, check succeeded).
export function countUpdatable(mods: Mod[], checks: UpdateChecksMap): number {
  return mods.filter(m => {
    const c = checks.get(m.id)
    return !!c && c.updateAvailable && !c.error && !m.pinned
  }).length
}

export function updatesAvailableText(count: number): string {
  if (count <= 0) return "All mods up to date"
  return `${count} ${count === 1 ? "update" : "updates"} available`
}

export function updateAllLabel(count: number): string {
  return count > 0 ? `Update All (${count})` : "Update All"
}

export function formatCheckedAgo(checkedAt: string | null | undefined, now: number = Date.now()): string {
  const t = Date.parse(checkedAt ?? "")
  if (Number.isNaN(t)) return ""
  const minutes = Math.floor(Math.max(0, now - t) / 60000)
  if (minutes < 1) return "Checked just now"
  if (minutes < 60) return `Checked ${minutes} min ago`
  const hours = Math.floor(minutes / 60)
  if (hours < 24) return `Checked ${hours} h ago`
  return `Checked ${Math.floor(hours / 24)} d ago`
}

export type ModUpdateBadgeKind = "available" | "pinned" | "error"

export interface ModUpdateBadge {
  kind: ModUpdateBadgeKind
  tooltip: string
}

// What (if anything) a mod card shows for its check result.
export function modUpdateBadge(mod: Pick<Mod, "pinned">, check: UpdateCheckItem | undefined): ModUpdateBadge | null {
  if (!check) return null
  if (check.error) return {kind: "error", tooltip: `Update check failed: ${check.error}`}
  if (!check.updateAvailable) return null
  const tooltip = check.updateString || "A newer version is available"
  return {kind: mod.pinned ? "pinned" : "available", tooltip}
}

// Ids of mods whose check found an update (pinned included: they still have one).
export function updateAvailableIds(checks: UpdateChecksMap): Set<number> {
  const ids = new Set<number>()
  for (const [id, c] of checks) if (c.updateAvailable && !c.error) ids.add(id)
  return ids
}
