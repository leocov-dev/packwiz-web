import {describe, expect, it} from "vitest"
import {
  filterItems, groupCounts, initialReviewSource, isActiveBulkStatus, isReviewable, itemGroup,
  parseBulkLines, progressText, replaceItem, sourceLabel, statusMeta,
} from "@/lib/bulk-add.ts"
import type {BulkAddItem, BulkItemStatus} from "@/interfaces/pack.ts"

const item = (id: number, status: BulkItemStatus, o: Partial<BulkAddItem> = {}): BulkAddItem =>
  ({id, position: id, line: `mod${id}`, status, source: "", slug: "", name: "", projectId: "",
    reason: "", conflictsWithPosition: 0, candidates: [], ...o}) as BulkAddItem

describe("parseBulkLines", () => {
  it("trims, drops blanks, comments and repeats, keeps order", () => {
    expect(parseBulkLines("  Sodium \r\n\n# note\nJust   Enough Items\nsodium\nlithium"))
      .toEqual(["Sodium", "Just Enough Items", "lithium"])
  })
  it("returns nothing for empty text", () => {
    expect(parseBulkLines("\n  \n# only\n")).toEqual([])
  })
})

describe("statuses", () => {
  it("knows active runs", () => {
    expect(isActiveBulkStatus("queued")).toBe(true)
    expect(isActiveBulkStatus("running")).toBe(true)
    expect(isActiveBulkStatus("done")).toBe(false)
    expect(isActiveBulkStatus(undefined)).toBe(false)
  })
  it("knows reviewable lines", () => {
    expect(isReviewable("pending")).toBe(true)
    expect(isReviewable("conflict")).toBe(true)
    expect(isReviewable("incompatible_version")).toBe(true)
    expect(isReviewable("added")).toBe(false)
    expect(isReviewable("rejected")).toBe(false)
  })
  it("groups lines", () => {
    expect(itemGroup("picked")).toBe("added")
    expect(itemGroup("rejected")).toBe("skipped")
    expect(itemGroup("conflict")).toBe("review")
    expect(itemGroup("failed")).toBe("failed")
    expect(itemGroup("queued")).toBe("queued")
  })
  it("labels every status", () => {
    const all: BulkItemStatus[] = ["queued", "added", "pending", "incompatible_version", "conflict", "skipped", "failed", "picked", "rejected"]
    for (const s of all) expect(statusMeta(s).label).not.toBe(s)
    expect(statusMeta("failed").color).toBe("error")
    expect(statusMeta("skipped").color).toBeUndefined()
  })
})

describe("groupCounts / filterItems", () => {
  const items = [item(1, "added"), item(2, "pending"), item(3, "conflict"), item(4, "skipped"), item(5, "failed"), item(6, "picked")]
  it("counts per group", () => {
    expect(groupCounts(items)).toEqual({all: 6, review: 2, added: 2, skipped: 1, failed: 1})
  })
  it("filters in list order", () => {
    expect(filterItems(items, "review").map(i => i.id)).toEqual([2, 3])
    expect(filterItems(items, "all")).toHaveLength(6)
  })
})

describe("initialReviewSource", () => {
  it("prefers the matched source, then the first candidate", () => {
    expect(initialReviewSource({source: "curseforge", candidates: []}, true)).toBe("curseforge")
    expect(initialReviewSource({source: "", candidates: [{source: "curseforge", slug: "a", title: "A", projectId: "1"}]}, true)).toBe("curseforge")
    expect(initialReviewSource({source: "", candidates: []}, true)).toBe("modrinth")
  })
  it("falls back to modrinth without a CurseForge key", () => {
    expect(initialReviewSource({source: "curseforge", candidates: []}, false)).toBe("modrinth")
  })
})

describe("helpers", () => {
  it("replaces an item by id", () => {
    const out = replaceItem([item(1, "pending"), item(2, "pending")], item(2, "picked"))
    expect(out.map(i => i.status)).toEqual(["pending", "picked"])
  })
  it("labels sources", () => {
    expect(sourceLabel("modrinth")).toBe("Modrinth")
    expect(sourceLabel("curseforge")).toBe("CurseForge")
    expect(sourceLabel("")).toBe("")
  })
  it("describes progress", () => {
    expect(progressText("running", 1)).toBe("Looking up 1 mod...")
    expect(progressText("queued", 3)).toBe("Waiting to look up 3 mods...")
  })
})
