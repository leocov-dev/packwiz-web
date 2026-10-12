import "reflect-metadata"
import {afterEach, beforeEach, describe, expect, it, vi} from "vitest"
import type {CompatResponse} from "@/interfaces/compat.ts"

const {fetchCompat, requestCompatScan} = vi.hoisted(() => ({fetchCompat: vi.fn(), requestCompatScan: vi.fn()}))
vi.mock("@/services/compat.service.ts", () => ({fetchCompat, requestCompatScan}))

import {useCompat} from "@/composables/useCompat.ts"
import {COMPAT_POLL_MAX_DURATION_MS, COMPAT_POLL_MS} from "@/lib/compat.ts"

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
})
