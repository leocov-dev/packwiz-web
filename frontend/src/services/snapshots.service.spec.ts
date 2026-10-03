import "reflect-metadata"
import {beforeEach, describe, expect, it, vi} from "vitest"

const {get, post} = vi.hoisted(() => ({get: vi.fn(), post: vi.fn()}))

vi.mock("@/services/api.service.ts", () => ({
  apiClient: {get, post},
}))

import {
  cloneFromSnapshot,
  fetchPackSnapshot,
  fetchPackSnapshots,
  revertToSnapshot,
} from "@/services/snapshots.service.ts"
import {PackSnapshot} from "@/interfaces/snapshot.ts"

beforeEach(() => {
  get.mockReset()
  post.mockReset()
})

const snapshotJson = {
  id: 7, seq: 3, parentId: 6, reason: "mod_pin", detail: {slug: "sodium"},
  summary: {added: 0, removed: 0, changed: 1, packFields: []},
  createdAt: "2026-10-03T10:00:00Z", createdBy: 1, createdByUsername: "admin",
  abandonedAt: null, abandonedByRevertTo: null, isHead: true,
}

describe("fetchPackSnapshots", () => {
  it("requests a page with the abandoned flag and hydrates snapshots", async () => {
    get.mockResolvedValue({data: {snapshots: [snapshotJson], total: 1, headId: 7, packStatus: "published"}})

    const result = await fetchPackSnapshots(12, 2, 25, true)

    expect(get).toHaveBeenCalledWith("v1/packwiz/pack/12/snapshots?page=2&pageSize=25&abandoned=true")
    expect(result.total).toBe(1)
    expect(result.snapshots[0]).toBeInstanceOf(PackSnapshot)
    expect(result.snapshots[0].isAbandoned).toBe(false)
  })

  it("flags an abandoned snapshot", async () => {
    get.mockResolvedValue({
      data: {
        snapshots: [{...snapshotJson, abandonedAt: "2026-10-03T11:00:00Z", abandonedByRevertTo: 2}],
        total: 1, headId: 2, packStatus: "draft",
      },
    })

    const result = await fetchPackSnapshots(12, 1, 25, false)

    expect(get).toHaveBeenCalledWith("v1/packwiz/pack/12/snapshots?page=1&pageSize=25&abandoned=false")
    expect(result.snapshots[0].isAbandoned).toBe(true)
  })
})

describe("fetchPackSnapshot", () => {
  it("passes the diff base", async () => {
    get.mockResolvedValue({
      data: {
        snapshot: snapshotJson, against: "current",
        diff: {pack: [], added: [], removed: [], changed: [{slug: "a", name: "A", version: "1", fileName: "a.jar", changes: [{field: "pinned", from: false, to: true}]}]},
      },
    })

    const result = await fetchPackSnapshot(12, 7, "current")

    expect(get).toHaveBeenCalledWith("v1/packwiz/pack/12/snapshots/7?against=current")
    expect(result.diff.changed[0].changes[0].field).toBe("pinned")
  })
})

describe("revertToSnapshot", () => {
  it("posts to the revert route", async () => {
    post.mockResolvedValue({data: {changed: true, headId: 7}})

    const result = await revertToSnapshot(12, 7)

    expect(post).toHaveBeenCalledWith("v1/packwiz/pack/12/snapshots/7/revert")
    expect(result).toMatchObject({changed: true, headId: 7})
  })
})

describe("cloneFromSnapshot", () => {
  it("posts the new name and slug and returns the new pack", async () => {
    post.mockResolvedValue({data: {id: 99, slug: "copy", name: "Copy", permissions: ["pack.view"]}})

    const pack = await cloneFromSnapshot(12, 7, {slug: "copy", name: "Copy"})

    expect(post).toHaveBeenCalledWith("v1/packwiz/pack/12/snapshots/7/clone", {slug: "copy", name: "Copy"})
    expect(pack.id).toBe(99)
  })
})
