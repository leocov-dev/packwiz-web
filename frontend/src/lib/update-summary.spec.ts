import {describe, expect, it} from "vitest"
import {skipReasonLabel, summarizeUpdateAll, updateAllOutcome} from "@/lib/update-summary.ts"
import type {UpdateAllItem, UpdateAllResponse} from "@/interfaces/pack.ts"

const item = (slug: string): UpdateAllItem => ({modId: 1, slug, name: slug})
const resp = (o: Partial<UpdateAllResponse> = {}): UpdateAllResponse =>
  ({updated: [], skipped: [], failed: [], upToDate: 0, notChecked: 0, ...o})

describe("updateAllOutcome", () => {
  it("is up-to-date when nothing to report", () => {
    expect(updateAllOutcome(resp({upToDate: 4}))).toBe("up-to-date")
  })
  it.each([
    ["updated", resp({updated: [item("a")]})],
    ["skipped", resp({skipped: [item("a")]})],
    ["failed", resp({failed: [item("a")]})],
  ])("has details when %s", (_n, r) => {
    expect(updateAllOutcome(r)).toBe("details")
  })
})

describe("summarizeUpdateAll", () => {
  it("reports all up to date", () => {
    expect(summarizeUpdateAll(resp({upToDate: 3}))).toBe("All mods are up to date")
  })
  it("reports unchecked mods honestly", () => {
    expect(summarizeUpdateAll(resp({upToDate: 12, notChecked: 3})))
      .toBe("12 up to date, 3 not checked (manual sources)")
  })
  it("singular update", () => {
    expect(summarizeUpdateAll(resp({updated: [item("a")]}))).toBe("1 mod updated")
  })
  it("combines sections", () => {
    expect(summarizeUpdateAll(resp({
      updated: [item("a"), item("b")],
      failed: [item("c")],
      skipped: [item("d")],
      upToDate: 5,
    }))).toBe("2 mods updated, 1 failed, 1 skipped, 5 up to date")
  })
})

describe("skipReasonLabel", () => {
  it.each([
    ["pinned", "Pinned"],
    ["something-new", "Something-new"],
    [undefined, ""],
  ])("%s -> %s", (reason, label) => {
    expect(skipReasonLabel(reason)).toBe(label)
  })
})
