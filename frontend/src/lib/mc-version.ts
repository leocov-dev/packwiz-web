// Mirrors backend utils.CompareMinecraftVersions.

const STAGES: Record<string, number> = {snapshot: 0, pre: 1, rc: 2}
const RELEASE = 3

interface McVersion {
  nums: number[]
  stage: number
  build: number
}

function parse(v: string): McVersion | null {
  const trimmed = v.trim()
  const dash = trimmed.indexOf("-")
  const core = dash < 0 ? trimmed : trimmed.slice(0, dash)
  const parts = core.split(".")
  if (parts.length < 2 || parts.some(p => !/^\d+$/.test(p))) return null

  const out: McVersion = {nums: parts.map(Number), stage: RELEASE, build: 0}
  if (dash < 0) return out

  // "snapshot-2", "pre-1", "rc-3" (year-based) or "pre1", "rc1" (older)
  const match = /^([a-z]+)(\d*)$/.exec(trimmed.slice(dash + 1).replace(/-/g, ""))
  if (!match || !(match[1]! in STAGES)) return null
  out.stage = STAGES[match[1]!]!
  out.build = match[2] ? Number(match[2]) : 0
  return out
}

/**
 * Orders two Minecraft version ids (-1 older, 0 same, 1 newer). Year-based
 * ids ("26.1") sort after "1.x". A dev build sorts before its release.
 * Returns null when either id can't be compared (e.g. "24w14a").
 */
export function compareMinecraftVersions(a: string, b: string): number | null {
  const va = parse(a)
  const vb = parse(b)
  if (!va || !vb) return null

  for (let i = 0; i < Math.max(va.nums.length, vb.nums.length); i++) {
    const d = (va.nums[i] ?? 0) - (vb.nums[i] ?? 0)
    if (d !== 0) return Math.sign(d)
  }
  if (va.stage !== vb.stage) return Math.sign(va.stage - vb.stage)
  return Math.sign(va.build - vb.build)
}
