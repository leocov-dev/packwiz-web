import type {PackSnapshot, SnapshotSummary} from "@/interfaces/snapshot.ts"

const REASON_LABELS: Record<string, string> = {
  baseline: "Baseline",
  publish: "Published",
  pack_edit: "Pack edited",
  mod_add: "Mod added",
  mod_remove: "Mod removed",
  mod_update: "Mod updated",
  mod_side: "Side changed",
  mod_option: "Option changed",
  mod_pin: "Pin changed",
  rehash: "Rehashed",
  update_all: "Updated all mods",
  migrate: "Migrated",
  migrate_mods: "Migration mod updates",
}

const REASON_ICONS: Record<string, string> = {
  baseline: "mdi-flag-outline",
  publish: "mdi-publish",
  pack_edit: "mdi-pencil",
  mod_add: "mdi-plus-circle-outline",
  mod_remove: "mdi-minus-circle-outline",
  mod_update: "mdi-update",
  mod_side: "mdi-swap-horizontal",
  mod_option: "mdi-tune",
  mod_pin: "mdi-pin-outline",
  rehash: "mdi-pound",
  update_all: "mdi-update",
  migrate: "mdi-arrow-up-bold-hexagon-outline",
  migrate_mods: "mdi-arrow-up-bold-hexagon-outline",
}

const PACK_FIELD_LABELS: Record<string, string> = {
  name: "Name",
  description: "Description",
  version: "Pack version",
  packFormat: "Pack format",
  mcVersion: "Minecraft version",
  loader: "Loader",
  loaderVersion: "Loader version",
  acceptableGameVersions: "Acceptable game versions",
}

const MOD_FIELD_LABELS: Record<string, string> = {
  name: "Name",
  fileName: "File name",
  side: "Side",
  pinned: "Pinned",
  hashFormat: "Hash format",
  alias: "Alias",
  type: "Type",
  source: "Source",
  preserve: "Preserve",
  "download.url": "Download URL",
  "download.mode": "Download mode",
  "download.hash": "Download hash",
  "download.hashFormat": "Download hash format",
  update: "Update data",
  "option.optional": "Optional",
  "option.description": "Option description",
  "option.default": "Enabled by default",
  isDependency: "Dependency",
  dependencies: "Depends on",
  version: "Installed version",
}

export type FieldScope = "pack" | "mod"

export function snapshotReasonLabel(reason: string): string {
  return REASON_LABELS[reason] ?? reason
}

export function snapshotReasonIcon(reason: string): string {
  return REASON_ICONS[reason] ?? "mdi-history"
}

export function fieldLabel(field: string, scope: FieldScope): string {
  const labels = scope === "pack" ? PACK_FIELD_LABELS : MOD_FIELD_LABELS
  return labels[field] ?? field
}

/** The mod a snapshot's change was about, when its detail names one. */
export function snapshotSubject(detail: Record<string, unknown> | null | undefined): string {
  if (!detail) {
    return ""
  }
  const name = detail.name
  if (typeof name === "string" && name !== "") {
    return name
  }
  const slug = detail.slug
  return typeof slug === "string" ? slug : ""
}

/** One line describing what a snapshot changed, e.g. "2 added · 1 changed · Minecraft version". */
export function summarizeSnapshot(summary: SnapshotSummary): string {
  const parts: string[] = []
  if (summary.added > 0) {
    parts.push(`${summary.added} added`)
  }
  if (summary.removed > 0) {
    parts.push(`${summary.removed} removed`)
  }
  if (summary.changed > 0) {
    parts.push(`${summary.changed} changed`)
  }
  for (const field of summary.packFields ?? []) {
    parts.push(fieldLabel(field, "pack"))
  }
  return parts.length > 0 ? parts.join(" · ") : "No changes"
}

export function formatFieldValue(value: unknown): string {
  if (value === null || value === undefined || value === "") {
    return "—"
  }
  if (typeof value === "boolean") {
    return value ? "Yes" : "No"
  }
  if (Array.isArray(value)) {
    return value.length > 0 ? value.map(v => String(v)).join(", ") : "—"
  }
  return String(value)
}

/** An abandoned snapshot can be viewed and cloned, never restored. */
export function canRevertSnapshot(snapshot: Pick<PackSnapshot, "isAbandoned">, hasRevertPermission: boolean): boolean {
  return hasRevertPermission && !snapshot.isAbandoned
}

/** Reverting to the head only discards changes made since it was recorded. */
export function revertLabel(snapshot: Pick<PackSnapshot, "isHead">): string {
  return snapshot.isHead ? "Discard unrecorded changes" : "Revert to this snapshot"
}

export function revertConfirmText(snapshot: Pick<PackSnapshot, "isHead" | "seq">): string {
  if (snapshot.isHead) {
    return `The pack will be restored to snapshot #${snapshot.seq}. Changes made since it was recorded, such as edits made while the pack was a draft, will be lost.`
  }
  return `The pack will be restored to snapshot #${snapshot.seq}. Every later snapshot will be marked abandoned and can no longer be restored, though it stays in the history. This cannot be undone.`
}

/** An abandoned snapshot is never part of the live history, so it cannot be a new root. */
export function canRebaseSnapshot(snapshot: Pick<PackSnapshot, "isAbandoned">, hasManagePermission: boolean): boolean {
  return hasManagePermission && !snapshot.isAbandoned
}

export function rebaseConfirmText(snapshot: Pick<PackSnapshot, "seq">): string {
  return `Snapshot #${snapshot.seq} becomes the first snapshot of this pack. Every snapshot before it, including abandoned ones, is permanently deleted and can no longer be viewed, cloned or restored. The pack itself is not changed. This cannot be undone.`
}

export const pruneConfirmText =
  "Every abandoned snapshot of this pack is permanently deleted and can no longer be viewed or cloned. Live history and the pack itself are not changed. This cannot be undone."
