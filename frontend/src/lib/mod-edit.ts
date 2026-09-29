import type {Mod} from "@/interfaces/pack.ts"

export interface ModEditValues {
  side: Mod["side"]
  pinned: boolean
  optional: boolean
  description: string
  default: boolean
}

export interface ModEditDiff {
  side: boolean
  pinned: boolean
  option: boolean
  any: boolean
}

export const MAX_OPTION_DESCRIPTION = 500

export function editValuesFromMod(mod: Mod): ModEditValues {
  return {
    side: mod.side,
    pinned: !!mod.pinned,
    optional: mod.option?.optional ?? false,
    description: mod.option?.description ?? "",
    default: mod.option?.default ?? false,
  }
}

// Description/default are hidden while `optional` is off, so they only count
// as changes when the mod is (or was) optional.
export function diffModEdit(initial: ModEditValues, current: ModEditValues): ModEditDiff {
  const side = initial.side !== current.side
  const pinned = initial.pinned !== current.pinned
  const option = initial.optional !== current.optional
    || (current.optional
      && (initial.description !== current.description || initial.default !== current.default))
  return {side, pinned, option, any: side || pinned || option}
}

// What the option endpoint should receive: hidden fields are cleared when optional is off.
export function buildOptionRequest(values: ModEditValues) {
  return values.optional
    ? {optional: true, description: values.description, default: values.default}
    : {optional: false, description: "", default: false}
}

export function describeSaveFailure(saved: string[], failed: string, message: string, notAttempted: string[] = []): string {
  const parts: string[] = []
  if (saved.length > 0) parts.push(`${saved.join(", ")} saved.`)
  parts.push(`${failed} failed: ${message}.`)
  if (notAttempted.length > 0) parts.push(`Not saved: ${notAttempted.join(", ")}.`)
  return parts.join(" ")
}
