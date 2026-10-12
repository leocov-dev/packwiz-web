import "reflect-metadata"
import {beforeEach, describe, expect, it, vi} from "vitest"

const {get, post} = vi.hoisted(() => ({get: vi.fn(), post: vi.fn()}))

vi.mock("@/services/api.service", () => ({
  apiClient: {get, post},
}))

import {fetchCompat, requestCompatScan} from "@/services/compat.service.ts"
import {CompatFindingItem, CompatResponse, CompatScanResponse} from "@/interfaces/compat.ts"
import {AllPacksResponse} from "@/interfaces/pack.ts"
import {plainToInstance} from "class-transformer"

beforeEach(() => {
  get.mockReset()
  post.mockReset()
})

describe("fetchCompat", () => {
  it("gets the pack compat and hydrates findings", async () => {
    get.mockResolvedValue({
      data: {
        status: "errors",
        counts: {errors: 1, warnings: 0, info: 0},
        unscanned: 2,
        scannedAt: "2026-01-01T00:00:00Z",
        runError: "",
        findings: [{modId: 3, slug: "sodium", severity: "error", code: "wrong_loader", message: "m", detail: {a: "b"}}],
      },
    })
    const r = await fetchCompat(7)
    expect(get).toHaveBeenCalledWith("v1/packwiz/pack/7/compat")
    expect(r).toBeInstanceOf(CompatResponse)
    expect(r.findings[0]).toBeInstanceOf(CompatFindingItem)
    expect(r.findings[0].detail).toEqual({a: "b"})
    expect(r.unscanned).toBe(2)
  })
})

describe("requestCompatScan", () => {
  it("posts the scan request", async () => {
    post.mockResolvedValue({data: {status: "queued", enqueued: true}})
    const r = await requestCompatScan(7)
    expect(post).toHaveBeenCalledWith("v1/packwiz/pack/7/compat/scan")
    expect(r).toBeInstanceOf(CompatScanResponse)
    expect(r.enqueued).toBe(true)
  })

  it("propagates errors", async () => {
    post.mockRejectedValue(new Error("409"))
    await expect(requestCompatScan(7)).rejects.toThrow("409")
  })
})

describe("pack compat summary", () => {
  it("is optional on packs", () => {
    const r = plainToInstance(AllPacksResponse, {
      packs: [{id: 1, compat: {status: "warnings", errors: 0, warnings: 2}}, {id: 2}],
    })
    expect(r.packs?.[0].compat?.warnings).toBe(2)
    expect(r.packs?.[1].compat).toBeUndefined()
  })
})
