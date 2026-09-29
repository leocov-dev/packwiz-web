import {describe, expect, it} from "vitest"
import {
  CHECK_MIN_INTERVAL_MS, CHECK_TTL_MS, checkCooldownMs, shouldForceCheck, shouldGiveUpPolling,
  buildResultsMap, countUpdatable, formatCheckedAgo, modUpdateBadge, nextPollDelay,
  POLL_MAX_DELAY_MS, POLL_MAX_DURATION_MS, shouldContinuePolling, updateAvailableIds, updateAllLabel, updatesAvailableText,
} from "@/lib/update-checks.ts"
import type {Mod, UpdateCheckItem} from "@/interfaces/pack.ts"

const mod = (id: number, pinned = false) => ({id, pinned}) as Mod
const chk = (modId: number, o: Partial<UpdateCheckItem> = {}): UpdateCheckItem =>
  ({modId, updateAvailable: false, ...o})

describe("polling", () => {
  it("starts at 2s and caps", () => {
    expect(nextPollDelay(0)).toBe(2000)
    expect(nextPollDelay(1)).toBeGreaterThan(2000)
    expect(nextPollDelay(50)).toBe(POLL_MAX_DELAY_MS)
    expect(nextPollDelay(-3)).toBe(2000)
  })
  it("continues only while active and under the max duration", () => {
    expect(shouldContinuePolling("queued", 0)).toBe(true)
    expect(shouldContinuePolling("running", POLL_MAX_DURATION_MS - 1)).toBe(true)
    expect(shouldContinuePolling("running", POLL_MAX_DURATION_MS)).toBe(false)
    for (const s of ["idle", "done", "failed"] as const) expect(shouldContinuePolling(s, 0)).toBe(false)
  })
})

describe("results", () => {
  it("builds a map by mod id", () => {
    expect(buildResultsMap([chk(1), chk(2)]).size).toBe(2)
    expect(buildResultsMap(undefined).size).toBe(0)
  })
  it("counts only unpinned, error-free updates", () => {
    const map = buildResultsMap([
      chk(1, {updateAvailable: true}),
      chk(2, {updateAvailable: true}),
      chk(3, {updateAvailable: true, error: "x"}),
      chk(4),
    ])
    expect(countUpdatable([mod(1), mod(2, true), mod(3), mod(4), mod(5)], map)).toBe(1)
  })
})

describe("text", () => {
  it("formats availability and label", () => {
    expect(updatesAvailableText(0)).toBe("All mods up to date")
    expect(updatesAvailableText(1)).toBe("1 update available")
    expect(updatesAvailableText(3)).toBe("3 updates available")
    expect(updateAllLabel(0)).toBe("Update All")
    expect(updateAllLabel(4)).toBe("Update All (4)")
  })
  it("formats checked ago", () => {
    const now = Date.parse("2026-01-01T12:00:00Z")
    const ago = (ms: number) => new Date(now - ms).toISOString()
    expect(formatCheckedAgo(ago(10_000), now)).toBe("Checked just now")
    expect(formatCheckedAgo(ago(3 * 60_000), now)).toBe("Checked 3 min ago")
    expect(formatCheckedAgo(ago(2 * 3600_000), now)).toBe("Checked 2 h ago")
    expect(formatCheckedAgo(ago(50 * 3600_000), now)).toBe("Checked 2 d ago")
    expect(formatCheckedAgo(null, now)).toBe("")
    expect(formatCheckedAgo("junk", now)).toBe("")
  })
})

describe("modUpdateBadge", () => {
  it("covers each state", () => {
    expect(modUpdateBadge(mod(1), undefined)).toBeNull()
    expect(modUpdateBadge(mod(1), chk(1))).toBeNull()
    expect(modUpdateBadge(mod(1), chk(1, {updateAvailable: true, updateString: "a -> b"})))
      .toEqual({kind: "available", tooltip: "a -> b"})
    expect(modUpdateBadge(mod(1, true), chk(1, {updateAvailable: true}))?.kind).toBe("pinned")
    expect(modUpdateBadge({...mod(1), version: "1.2.3"}, chk(1, {updateAvailable: true, updateString: "a -> b", latestVersion: "1.3.0"})))
      .toEqual({kind: "available", tooltip: "1.2.3 -> 1.3.0"})
    // latest alone (no installed version) keeps the server string
    expect(modUpdateBadge(mod(1), chk(1, {updateAvailable: true, updateString: "a -> b", latestVersion: "1.3.0"})))
      .toEqual({kind: "available", tooltip: "a -> b"})
    expect(modUpdateBadge(mod(1), chk(1, {error: "rate limited"})))
      .toEqual({kind: "error", tooltip: "Update check failed: rate limited"})
  })
})

describe("updateAvailableIds", () => {
  it("collects mods with a successful available update", () => {
    const map = buildResultsMap([chk(1, {updateAvailable: true}), chk(2, {updateAvailable: true, error: "x"}), chk(3)])
    expect([...updateAvailableIds(map)]).toEqual([1])
  })
})

describe("poll failure tolerance", () => {
  it("tolerates up to 3 consecutive failures", () => {
    expect(shouldGiveUpPolling(1)).toBe(false)
    expect(shouldGiveUpPolling(3)).toBe(false)
    expect(shouldGiveUpPolling(4)).toBe(true)
  })
})

describe("check cooldown and force", () => {
  const now = Date.parse("2026-01-01T12:00:00Z")
  const ago = (ms: number) => new Date(now - ms).toISOString()

  it("cooldown counts down from the last run", () => {
    expect(checkCooldownMs(ago(10_000), now)).toBe(CHECK_MIN_INTERVAL_MS - 10_000)
    expect(checkCooldownMs(ago(CHECK_MIN_INTERVAL_MS), now)).toBe(0)
    expect(checkCooldownMs(null, now)).toBe(0)
    expect(checkCooldownMs("garbage", now)).toBe(0)
  })

  it("forces only after the cached result expired", () => {
    expect(shouldForceCheck("done", ago(60_000), now)).toBe(false)
    expect(shouldForceCheck("done", ago(CHECK_TTL_MS), now)).toBe(true)
    expect(shouldForceCheck("idle", ago(CHECK_TTL_MS * 2), now)).toBe(false)
    expect(shouldForceCheck("done", null, now)).toBe(false)
  })
})
