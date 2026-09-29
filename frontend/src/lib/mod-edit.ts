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

export interface SaveStep {
  label: string
  run: () => Promise<void>
}

export interface StepResult {
  label: string
  ok: boolean
  message?: string
}

// Steps touch independent columns, so every step is attempted even if an earlier one fails.
export async function runSaveSteps(
  steps: SaveStep[],
  toMessage: (e: unknown) => string = e => (e instanceof Error ? e.message : String(e)),
): Promise<StepResult[]> {
  const results: StepResult[] = []
  for (const step of steps) {
    try {
      await step.run()
      results.push({label: step.label, ok: true})
    } catch (e) {
      results.push({label: step.label, ok: false, message: toMessage(e)})
    }
  }
  return results
}

export function describeSaveFailure(results: StepResult[]): string {
  return results
    .map(r => (r.ok ? `${r.label} saved.` : `${r.label} failed: ${r.message}.`))
    .join(" ")
}

export function describeUpdateResult(name: string, updated: boolean, fileName: string): string {
  return updated
    ? `Updated ${name} to ${fileName}`
    : `${name} is already up to date`
}
