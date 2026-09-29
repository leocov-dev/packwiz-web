import type {Mod, UpdateCheckItem, UpdateCheckStatus} from "@/interfaces/pack.ts"
import {modVersion} from "@/lib/mod-filters.ts"

export type UpdateChecksMap = ReadonlyMap<number, UpdateCheckItem>

export const POLL_BASE_MS = 2000
export const POLL_MAX_DELAY_MS = 8000
export const POLL_MAX_DURATION_MS = 5 * 60 * 1000

// Mirrors the server: results younger than the TTL are served from cache, and
// no pack is re-checked within the minimum interval (even when forced).
export const CHECK_TTL_MS = 10 * 60 * 1000
export const CHECK_MIN_INTERVAL_MS = 60 * 1000
// Consecutive failed polls tolerated before the check is shown as failed.
export const MAX_POLL_FAILURES = 3

export const isActiveStatus = (status: UpdateCheckStatus): boolean =>
  status === "queued" || status === "running"

// ~2s at first, easing up to POLL_MAX_DELAY_MS.
export function nextPollDelay(attempt: number): number {
  return Math.min(POLL_BASE_MS * Math.pow(1.25, Math.max(0, attempt)), POLL_MAX_DELAY_MS)
}

export function shouldContinuePolling(status: UpdateCheckStatus, elapsedMs: number): boolean {
  return isActiveStatus(status) && elapsedMs < POLL_MAX_DURATION_MS
}

// True once too many consecutive poll requests failed (a single blip is tolerated).
export function shouldGiveUpPolling(consecutiveFailures: number): boolean {
  return consecutiveFailures > MAX_POLL_FAILURES
}

// Remaining ms before the server accepts another check (0 = allowed now).
export function checkCooldownMs(runFinishedAt: string | null | undefined, now: number = Date.now()): number {
  const t = Date.parse(runFinishedAt ?? "")
  if (Number.isNaN(t)) return 0
  return Math.max(0, CHECK_MIN_INTERVAL_MS - Math.max(0, now - t))
}

// Only force a re-check when the cached result has expired; within the TTL the
// server serves the cached result and force would just spend API rate limit.
export function shouldForceCheck(status: UpdateCheckStatus, checkedAt: string | null | undefined, now: number = Date.now()): boolean {
  if (status !== "done") return false
  const t = Date.parse(checkedAt ?? "")
  return !Number.isNaN(t) && now - t >= CHECK_TTL_MS
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
export function modUpdateBadge(mod: Pick<Mod, "pinned"> & Partial<Pick<Mod, "version">>, check: UpdateCheckItem | undefined): ModUpdateBadge | null {
  if (!check) return null
  if (check.error) return {kind: "error", tooltip: `Update check failed: ${check.error}`}
  if (!check.updateAvailable) return null
  // "1.2.3 -> 1.3.0" only when both versions are known (never the file-name fallback)
  const installed = modVersion(mod)
  const latest = (check.latestVersion ?? "").trim()
  const tooltip = installed && latest ? `${installed} -> ${latest}` : check.updateString || "A newer version is available"
  return {kind: mod.pinned ? "pinned" : "available", tooltip}
}

// Ids of mods whose check found an update (pinned included: they still have one).
export function updateAvailableIds(checks: UpdateChecksMap): Set<number> {
  const ids = new Set<number>()
  for (const [id, c] of checks) if (c.updateAvailable && !c.error) ids.add(id)
  return ids
}
