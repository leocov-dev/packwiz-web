import {formatAgo} from "@/lib/format-ago.ts"
import type {CompatFindingItem, CompatResponse, CompatSeverity, CompatStatus} from "@/interfaces/compat.ts"

export interface CompatDisplay {
  icon: string
  color: string
  /** Short chip text. */
  text: string
  /** Full sentence for tooltips and screen readers. */
  label: string
  variant: "flat" | "tonal" | "outlined"
  /** Show a spinner instead of the icon. */
  busy: boolean
}

export interface CompatCountsLike {
  errors: number
  warnings: number
  info?: number
}

export const COMPAT_POLL_MS = 3000
export const COMPAT_POLL_MAX_DURATION_MS = 3 * 60 * 1000

export const isCompatActive = (status: CompatStatus | undefined): boolean =>
  status === "queued" || status === "scanning"

const plural = (n: number, word: string) => `${n} ${word}${n === 1 ? "" : "s"}`

/** "2 errors, 1 warning"; empty when there is nothing to count. */
export function compatCountsText(counts: CompatCountsLike | undefined): string {
  if (!counts) return ""
  const parts: string[] = []
  if (counts.errors > 0) parts.push(plural(counts.errors, "error"))
  if (counts.warnings > 0) parts.push(plural(counts.warnings, "warning"))
  return parts.join(", ")
}

export function compatDisplay(status: CompatStatus, counts?: CompatCountsLike): CompatDisplay {
  const countsText = compatCountsText(counts)
  switch (status) {
    case "ok":
      return {icon: "mdi-check-circle", color: "success", text: "Compatible", label: "No compatibility problems found", variant: "tonal", busy: false}
    case "warnings":
      return {icon: "mdi-alert", color: "warning", text: countsText || "Warnings", label: `Compatibility warnings: ${countsText || "see details"}`, variant: "tonal", busy: false}
    case "errors":
      return {icon: "mdi-alert-circle", color: "error", text: countsText || "Errors", label: `Compatibility errors: ${countsText || "see details"}`, variant: "tonal", busy: false}
    case "queued":
      return {icon: "mdi-timer-sand", color: "info", text: "Queued", label: "Compatibility scan queued", variant: "tonal", busy: true}
    case "scanning":
      return {icon: "mdi-magnify-scan", color: "info", text: "Scanning", label: "Compatibility scan in progress", variant: "tonal", busy: true}
    case "stale":
      return {icon: "mdi-history", color: "grey", text: "Out of date", label: "Compatibility results are out of date", variant: "tonal", busy: false}
    case "failed":
      return {icon: "mdi-alert-octagon-outline", color: "error", text: "Scan failed", label: "Compatibility scan failed", variant: "outlined", busy: false}
    case "unknown":
    default:
      return {icon: "mdi-help-circle-outline", color: "grey", text: "Not scanned", label: "Compatibility has not been checked yet", variant: "tonal", busy: false}
  }
}

const FINDING_TITLES: Record<string, string> = {
  wrong_loader: "Wrong mod loader",
  mc_version_mismatch: "Minecraft version mismatch",
  loader_version_mismatch: "Loader version mismatch",
  missing_dependency: "Missing dependency",
  dependency_version: "Dependency version mismatch",
  breaks: "Incompatible mod installed",
  duplicate_mod_id: "Duplicate mod",
  conflicts: "Conflicting mod",
  parse_error: "Could not read mod metadata",
  no_metadata: "No mod metadata",
  not_scanned: "Not checked",
  side_mismatch: "Side mismatch",
  unevaluable_range: "Version range not evaluated",
}

/** Friendly title of a finding code; unknown codes show as they are. */
export function findingTitle(code: string): string {
  return FINDING_TITLES[code] ?? code
}

export const SEVERITY_ORDER: readonly CompatSeverity[] = ["error", "warning", "info"]

const SEVERITY_DISPLAY: Record<CompatSeverity, {title: string; icon: string; color: string}> = {
  error: {title: "Errors", icon: "mdi-alert-circle", color: "error"},
  warning: {title: "Warnings", icon: "mdi-alert", color: "warning"},
  info: {title: "Info", icon: "mdi-information", color: "info"},
}

export const severityDisplay = (s: CompatSeverity) => SEVERITY_DISPLAY[s]

export interface CompatModGroup {
  key: string
  slug: string
  modId: number | null
  findings: CompatFindingItem[]
}

export interface CompatSeverityGroup {
  severity: CompatSeverity
  mods: CompatModGroup[]
  total: number
}

/** Findings grouped by severity (errors first), then by mod; empty severities are dropped. */
export function groupFindings(findings: readonly CompatFindingItem[]): CompatSeverityGroup[] {
  const groups: CompatSeverityGroup[] = []
  for (const severity of SEVERITY_ORDER) {
    const mods = new Map<string, CompatModGroup>()
    let total = 0
    for (const f of findings) {
      if (f.severity !== severity) continue
      total++
      const key = f.modId !== null ? `id:${f.modId}` : `slug:${f.slug}`
      let g = mods.get(key)
      if (!g) {
        g = {key, slug: f.slug, modId: f.modId, findings: []}
        mods.set(key, g)
      }
      g.findings.push(f)
    }
    if (total > 0) groups.push({severity, mods: [...mods.values()], total})
  }
  return groups
}

export function compatUnscannedText(unscanned: number): string {
  return unscanned > 0 ? `${plural(unscanned, "mod")} could not be checked` : ""
}

export function formatScannedAgo(scannedAt: string | null | undefined, now: number = Date.now()): string {
  return formatAgo("Last scanned", scannedAt, now) || "Never scanned"
}

/** Pack facts that help a reviewer judge the findings. */
export interface CompatReportContext {
  name?: string
  mcVersion?: string
  loader?: string
  loaderVersion?: string
  /** pack.toml link (public or personalized). */
  packLink?: string
}

/** Clean plaintext report of a compat result (`modUrls`: source page per mod slug), meant to be pasted to an AI agent for review. */
export function formatCompatReport(
  compat: Pick<CompatResponse, "status" | "counts" | "unscanned" | "scannedAt" | "runError" | "findings">,
  context: CompatReportContext = {},
  modUrls: ReadonlyMap<string, string> = new Map(),
): string {
  const lines: string[] = ["Minecraft mod compatibility report"]
  if (context.name) lines.push(`Pack: ${context.name}`)
  if (context.mcVersion) lines.push(`Minecraft version: ${context.mcVersion}`)
  if (context.loader) {
    lines.push(`Loader: ${[context.loader, context.loaderVersion].filter(Boolean).join(" ")}`)
  }
  if (context.packLink) lines.push(`pack.toml: ${context.packLink}`)
  lines.push(`Status: ${compat.status}`)
  if (compat.scannedAt) lines.push(`Scanned at: ${compat.scannedAt}`)
  const countsText = compatCountsText(compat.counts)
  if (countsText) lines.push(`Findings: ${countsText}`)
  const unscannedText = compatUnscannedText(compat.unscanned)
  if (unscannedText) lines.push(unscannedText)
  if (compat.runError) lines.push(`Last scan problem: ${compat.runError}`)
  lines.push("Note: based on mod metadata only; does not guarantee the pack launches.")

  for (const group of groupFindings(compat.findings)) {
    lines.push("", `## ${severityDisplay(group.severity).title} (${group.total})`)
    for (const mod of group.mods) {
      lines.push("", mod.slug)
      const url = modUrls.get(mod.slug)
      if (url) lines.push(`Source: ${url}`)
      for (const finding of mod.findings) {
        lines.push(`- ${findingTitle(finding.code)} [${finding.code}]: ${finding.message}`)
        for (const [key, value] of Object.entries(finding.detail)) {
          lines.push(`    ${key}: ${value}`)
        }
      }
    }
  }
  if (compat.findings.length === 0) lines.push("", "No problems found.")
  return lines.join("\n") + "\n"
}

export function shouldContinueCompatPolling(status: CompatStatus, elapsedMs: number): boolean {
  return isCompatActive(status) && elapsedMs < COMPAT_POLL_MAX_DURATION_MS
}

/** Whether the user may start a scan: needs the permission, and not on an archived pack. */
export function canRescan(hasPermission: boolean, archived: boolean): boolean {
  return hasPermission && !archived
}

export const COMPAT_STALLED_MESSAGE = "The scan is taking too long. Reload the page to check again."

/** Status to show once polling is abandoned: never a busy one, so nothing spins forever. */
export const abandonedStatus = (status: CompatStatus): CompatStatus =>
  isCompatActive(status) ? "stale" : status
