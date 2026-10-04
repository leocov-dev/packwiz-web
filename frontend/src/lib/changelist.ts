import type {ChangelistEntry, ChangelistFieldChange, ChangelistPeriod} from "@/interfaces/changelist.ts"

function utcDate(date: string): Date | null {
  const [year, month, day] = date.split("-").map(Number)
  if (!year || !month || !day) return null
  return new Date(Date.UTC(year, month - 1, day))
}

function fmt(d: Date, opts: Intl.DateTimeFormatOptions): string {
  return d.toLocaleDateString("en-US", {...opts, timeZone: "UTC"})
}

function lastDayOfMonth(d: Date): number {
  return new Date(Date.UTC(d.getUTCFullYear(), d.getUTCMonth() + 1, 0)).getUTCDate()
}

function range(start: Date, end: Date): string {
  const sameYear = start.getUTCFullYear() === end.getUTCFullYear()
  const from = fmt(start, {month: "short", day: "numeric", ...(sameYear ? {} : {year: "numeric"})})
  return `${from} – ${fmt(end, {month: "short", day: "numeric", year: "numeric"})}`
}

/** Heading of an entry: "Sat, Oct 4", "September 2026", "2025", or a partial range. */
export function entryTitle(entry: ChangelistEntry): string {
  if (entry.period === "pending") return "Unpublished changes"
  const start = utcDate(entry.start)
  const end = utcDate(entry.end)
  if (!start || !end) return entry.start

  switch (entry.period) {
    case "day":
      return fmt(start, {weekday: "short", month: "short", day: "numeric", year: "numeric"})
    case "month":
      return end.getUTCDate() === lastDayOfMonth(end)
        ? fmt(start, {month: "long", year: "numeric"})
        : range(start, end)
    case "year":
      return end.getUTCMonth() === 11 && end.getUTCDate() === 31
        ? String(start.getUTCFullYear())
        : range(start, end)
  }
}

const SECTION_TITLES: Record<ChangelistPeriod, string> = {
  day: "Last 10 days",
  month: "Recent months",
  year: "Earlier",
  pending: "Unpublished",
}

export interface ChangelistSection {
  title: string
  entries: ChangelistEntry[]
}

/** Groups entries (newest first) under their period's heading. */
export function changelistSections(entries: readonly ChangelistEntry[]): ChangelistSection[] {
  const sections: ChangelistSection[] = []
  for (const entry of entries) {
    const title = SECTION_TITLES[entry.period]
    const last = sections[sections.length - 1]
    if (last?.title === title) {
      last.entries.push(entry)
    } else {
      sections.push({title, entries: [entry]})
    }
  }
  return sections
}

const PACK_FIELDS: Record<string, string> = {
  name: "Name",
  version: "Pack version",
  mcVersion: "Minecraft",
  loader: "Loader",
  loaderVersion: "Loader version",
  packFormat: "Pack format",
  acceptableGameVersions: "Accepted game versions",
}

function show(v: unknown): string {
  if (Array.isArray(v)) return v.length ? v.join(", ") : "none"
  if (v === undefined || v === null || v === "") return "none"
  return String(v)
}

export interface PackChangeLine {
  label: string
  /** "a → b", or empty when only the fact of a change is shown. */
  detail: string
  /** Changes what clients' instances run on (Minecraft or loader). */
  target: boolean
}

/** A pack-level change as one readable line. */
export function packChangeLine(change: ChangelistFieldChange, initial = false): PackChangeLine {
  const target = ["mcVersion", "loader", "loaderVersion"].includes(change.field)
  if (change.field === "description") {
    return {label: "Description updated", detail: "", target: false}
  }
  const label = PACK_FIELDS[change.field] ?? change.field
  return {label, detail: initial ? show(change.to) : `${show(change.from)} → ${show(change.to)}`, target}
}

/** Whether the entry changes the Minecraft version or loader type. */
export function changesTarget(entry: ChangelistEntry): boolean {
  return !entry.initial && entry.pack.some(c => c.field === "mcVersion" || c.field === "loader")
}

const MOD_FIELDS: Record<string, string> = {
  name: "renamed",
  side: "side changed",
  "option.optional": "optional changed",
  "option.default": "default changed",
}

export function modFieldLabel(field: string): string {
  return MOD_FIELDS[field] ?? field
}

/** Whether there's anything at all to show. */
export function entryIsEmpty(entry: ChangelistEntry): boolean {
  return !entry.initial && entry.pack.length === 0 && entry.added.length === 0
    && entry.removed.length === 0 && entry.changed.length === 0
}
