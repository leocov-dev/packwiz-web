import "reflect-metadata"
import {afterEach, beforeEach, describe, expect, it, vi} from "vitest"
import type {CompatResponse} from "@/interfaces/compat.ts"

const {fetchCompat, requestCompatScan} = vi.hoisted(() => ({fetchCompat: vi.fn(), requestCompatScan: vi.fn()}))
vi.mock("@/services/compat.service.ts", () => ({fetchCompat, requestCompatScan}))

import {useCompat} from "@/composables/useCompat.ts"
import {COMPAT_POLL_MAX_DURATION_MS, COMPAT_POLL_MS, COMPAT_STALLED_MESSAGE} from "@/lib/compat.ts"

const resp = (status: string, errors = 0): CompatResponse => ({
  status, counts: {errors, warnings: 0, info: 0}, unscanned: 0, scannedAt: null, runError: "", findings: [],
} as CompatResponse)

beforeEach(() => {
  vi.useFakeTimers()
  fetchCompat.mockReset()
  requestCompatScan.mockReset()
})
afterEach(() => vi.useRealTimers())

describe("useCompat", () => {
  it("does not fetch when the pack has no active summary", () => {
    const c = useCompat(() => 1, () => ({status: "ok", errors: 0, warnings: 0}))
    vi.advanceTimersByTime(COMPAT_POLL_MS * 3)
    expect(fetchCompat).not.toHaveBeenCalled()
    expect(c.status.value).toBe("ok")
  })

  it("polls an active summary until it settles", async () => {
    fetchCompat.mockResolvedValueOnce(resp("scanning")).mockResolvedValueOnce(resp("errors", 2))
    const c = useCompat(() => 1, () => ({status: "queued", errors: 0, warnings: 0}))
    await vi.advanceTimersByTimeAsync(COMPAT_POLL_MS)
    expect(c.status.value).toBe("scanning")
    await vi.advanceTimersByTimeAsync(COMPAT_POLL_MS)
    expect(c.status.value).toBe("errors")
    expect(c.summary.value?.errors).toBe(2)
    await vi.advanceTimersByTimeAsync(COMPAT_POLL_MS * 5)
    expect(fetchCompat).toHaveBeenCalledTimes(2)
  })

  it("stop() ends polling", async () => {
    fetchCompat.mockResolvedValue(resp("scanning"))
    const c = useCompat(() => 1, () => ({status: "queued", errors: 0, warnings: 0}))
    await vi.advanceTimersByTimeAsync(COMPAT_POLL_MS)
    c.stop()
    await vi.advanceTimersByTimeAsync(COMPAT_POLL_MS * 5)
    expect(fetchCompat).toHaveBeenCalledTimes(1)
  })

  it("gives up after the max duration", async () => {
    fetchCompat.mockResolvedValue(resp("scanning"))
    const c = useCompat(() => 1, () => ({status: "queued", errors: 0, warnings: 0}))
    await vi.advanceTimersByTimeAsync(COMPAT_POLL_MAX_DURATION_MS + COMPAT_POLL_MS * 2)
    const calls = fetchCompat.mock.calls.length
    await vi.advanceTimersByTimeAsync(COMPAT_POLL_MS * 5)
    expect(fetchCompat).toHaveBeenCalledTimes(calls)
    expect(c.error.value).not.toBe("")
  })

  it("rescan queues then polls", async () => {
    requestCompatScan.mockResolvedValue({status: "queued", enqueued: true})
    fetchCompat.mockResolvedValue(resp("ok"))
    const c = useCompat(() => 1, () => undefined)
    await c.rescan()
    expect(c.status.value).toBe("queued")
    await vi.advanceTimersByTimeAsync(COMPAT_POLL_MS)
    expect(c.status.value).toBe("ok")
  })

  it("rescan failure sets error", async () => {
    requestCompatScan.mockRejectedValue(new Error("boom"))
    const c = useCompat(() => 1, () => undefined)
    await c.rescan()
    expect(c.error.value).toBe("boom")
    expect(c.rescanning.value).toBe(false)
  })

  it("load fills the full result", async () => {
    fetchCompat.mockResolvedValue(resp("warnings"))
    const c = useCompat(() => 1, () => undefined)
    await c.load()
    expect(c.compat.value?.status).toBe("warnings")
    expect(c.loading.value).toBe(false)
  })

  it("sync resets to a new summary and resumes polling", async () => {
    fetchCompat.mockResolvedValue(resp("ok"))
    const c = useCompat(() => 1, () => ({status: "ok", errors: 0, warnings: 0}))
    c.sync({status: "queued", errors: 0, warnings: 0})
    await vi.advanceTimersByTimeAsync(COMPAT_POLL_MS)
    expect(fetchCompat).toHaveBeenCalledTimes(1)
    expect(c.status.value).toBe("ok")
  })

  const deferred = <T>() => {
    let resolve!: (v: T) => void
    let reject!: (e: unknown) => void
    const promise = new Promise<T>((res, rej) => { resolve = res; reject = rej })
    return {promise, resolve, reject}
  }

  it("loading clears when a load is superseded by rescan", async () => {
    const d = deferred<CompatResponse>()
    fetchCompat.mockReturnValueOnce(d.promise)
    requestCompatScan.mockResolvedValue({status: "queued", enqueued: true})
    const c = useCompat(() => 1, () => undefined)
    const loadP = c.load()
    expect(c.loading.value).toBe(true)
    await c.rescan()
    expect(c.loading.value).toBe(false)
    d.resolve(resp("ok"))
    await loadP
    expect(c.loading.value).toBe(false)
  })

  it("loading clears when load hands over to polling", async () => {
    fetchCompat.mockResolvedValueOnce(resp("scanning"))
    const c = useCompat(() => 1, () => undefined)
    await c.load()
    expect(c.status.value).toBe("scanning")
    expect(c.loading.value).toBe(false)
  })

  it("ignores a stale poll response after sync adopted newer state", async () => {
    const d = deferred<CompatResponse>()
    fetchCompat.mockReturnValueOnce(d.promise)
    const c = useCompat(() => 1, () => ({status: "queued", errors: 0, warnings: 0}))
    await vi.advanceTimersByTimeAsync(COMPAT_POLL_MS)
    c.sync({status: "ok", errors: 0, warnings: 0})
    d.resolve(resp("errors", 5))
    await vi.advanceTimersByTimeAsync(0)
    expect(c.status.value).toBe("ok")
    expect(c.summary.value?.errors).toBe(0)
    expect(c.compat.value).toBeNull()
  })

  it("ignores a stale load response after a newer load", async () => {
    const first = deferred<CompatResponse>()
    fetchCompat.mockReturnValueOnce(first.promise).mockResolvedValueOnce(resp("ok"))
    const c = useCompat(() => 1, () => undefined)
    const p1 = c.load()
    await c.load()
    first.resolve(resp("errors", 3))
    await p1
    expect(c.status.value).toBe("ok")
  })

  it("rescan racing an in-flight load keeps the rescan state", async () => {
    const d = deferred<CompatResponse>()
    fetchCompat.mockReturnValueOnce(d.promise).mockResolvedValue(resp("ok"))
    requestCompatScan.mockResolvedValue({status: "queued", enqueued: true})
    const c = useCompat(() => 1, () => ({status: "errors", errors: 2, warnings: 0}))
    const loadP = c.load()
    await c.rescan()
    d.resolve(resp("errors", 2))
    await loadP
    expect(c.status.value).toBe("queued")
    expect(c.summary.value?.errors).toBe(2)
    expect(c.rescanning.value).toBe(false)
    await vi.advanceTimersByTimeAsync(COMPAT_POLL_MS)
    expect(c.status.value).toBe("ok")
  })

  it("rescan keeps counts coherent in the loaded result", async () => {
    fetchCompat.mockResolvedValueOnce(resp("errors", 2))
    requestCompatScan.mockResolvedValue({status: "queued", enqueued: true})
    const c = useCompat(() => 1, () => undefined)
    await c.load()
    await c.rescan()
    expect(c.status.value).toBe("queued")
    expect(c.compat.value?.status).toBe("queued")
    expect(c.compat.value?.counts.errors).toBe(2)
    expect(c.summary.value?.errors).toBe(2)
  })

  it("rescan resolves to the failure message", async () => {
    requestCompatScan.mockRejectedValue(new Error("boom"))
    const c = useCompat(() => 1, () => undefined)
    expect(await c.rescan()).toBe("boom")
  })

  it("moves to a non-busy status when the max duration is hit", async () => {
    fetchCompat.mockResolvedValue(resp("scanning", 1))
    const c = useCompat(() => 1, () => ({status: "queued", errors: 1, warnings: 0}))
    await vi.advanceTimersByTimeAsync(COMPAT_POLL_MAX_DURATION_MS + COMPAT_POLL_MS * 2)
    expect(c.status.value).toBe("stale")
    expect(c.isActive.value).toBe(false)
    expect(c.summary.value?.errors).toBe(1)
    expect(c.error.value).toBe(COMPAT_STALLED_MESSAGE)
  })

  it("moves to a non-busy status after too many poll failures", async () => {
    fetchCompat.mockRejectedValue(new Error("down"))
    const c = useCompat(() => 1, () => ({status: "scanning", errors: 0, warnings: 1}))
    await vi.advanceTimersByTimeAsync(COMPAT_POLL_MS * 10)
    expect(c.status.value).toBe("stale")
    expect(c.summary.value?.warnings).toBe(1)
    expect(c.error.value).toBe("down")
  })
})
