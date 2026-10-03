import type {AccessDay} from "@/interfaces/access.ts";

export type AccessKind = "success" | "failure"

/** The per-day counts of one kind, oldest first, for charting. */
export function seriesValues(series: readonly AccessDay[], kind: AccessKind): number[] {
  return series.map((day) => day[kind])
}

export function seriesTotal(series: readonly AccessDay[], kind: AccessKind): number {
  return series.reduce((sum, day) => sum + day[kind], 0)
}

/** Formats a YYYY-MM-DD day as "Oct 3", independent of the viewer's time zone. */
export function formatDay(date: string): string {
  const [year, month, day] = date.split("-").map(Number)
  if (!year || !month || !day) return date
  return new Date(Date.UTC(year, month - 1, day)).toLocaleDateString("en-US", {
    month: "short",
    day: "numeric",
    timeZone: "UTC",
  })
}

/** Chart labels for a series; only every `step`th day is named so they stay legible. */
export function seriesLabels(series: readonly AccessDay[], step: number): string[] {
  return series.map((day, i) => (i % step === 0 ? formatDay(day.date) : ""))
}

/** Label spacing that keeps roughly 7 labels on screen. */
export function labelStep(dayCount: number): number {
  return Math.max(1, Math.ceil(dayCount / 7))
}

/** Adds an alpha channel to a #rgb or #rrggbb theme color for canvas fills. Other formats pass through. */
export function withAlpha(color: string, alpha: number): string {
  const match = /^#([0-9a-f]{3}|[0-9a-f]{6})$/i.exec(color.trim())
  if (!match) return color
  let hex = match[1]!
  if (hex.length === 3) hex = hex.split("").map((c) => c + c).join("")
  const n = parseInt(hex, 16)
  return `rgba(${(n >> 16) & 255}, ${(n >> 8) & 255}, ${n & 255}, ${alpha})`
}
