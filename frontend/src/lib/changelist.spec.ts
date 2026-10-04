import {describe, expect, it} from "vitest"
import type {ChangelistEntry} from "@/interfaces/changelist.ts"
import {changelistSections, changesTarget, entryIsEmpty, entryTitle, packChangeLine} from "./changelist.ts"

function entry(over: Partial<ChangelistEntry>): ChangelistEntry {
  return {
    period: "day", start: "2026-10-04", end: "2026-10-04", initial: false, modCount: 0,
    pack: [], added: [], removed: [], changed: [], ...over,
  }
}

describe("entryTitle", () => {
  it.each([
    [entry({}), "Sun, Oct 4, 2026"],
    [entry({period: "month", start: "2026-08-01", end: "2026-08-31"}), "August 2026"],
    [entry({period: "month", start: "2026-09-01", end: "2026-09-24"}), "Sep 1 – Sep 24, 2026"],
    [entry({period: "year", start: "2025-01-01", end: "2025-12-31"}), "2025"],
    [entry({period: "year", start: "2026-01-01", end: "2026-07-31"}), "Jan 1 – Jul 31, 2026"],
    [entry({period: "pending", start: "", end: ""}), "Unpublished changes"],
  ])("%#", (e, want) => {
    expect(entryTitle(e)).toBe(want)
  })
})

describe("changelistSections", () => {
  it("groups consecutive entries by period", () => {
    const sections = changelistSections([
      entry({}), entry({start: "2026-10-01"}),
      entry({period: "month"}), entry({period: "year"}),
    ])
    expect(sections.map(s => [s.title, s.entries.length])).toEqual([
      ["Last 10 days", 2], ["Recent months", 1], ["Earlier", 1],
    ])
  })
})

describe("packChangeLine", () => {
  it("shows from → to and flags target changes", () => {
    expect(packChangeLine({field: "mcVersion", from: "1.21.1", to: "26.1"}))
      .toEqual({label: "Minecraft", detail: "1.21.1 → 26.1", target: true})
  })
  it("hides description text", () => {
    expect(packChangeLine({field: "description", from: "", to: ""}).detail).toBe("")
  })
  it("shows only the value for an initial entry", () => {
    expect(packChangeLine({field: "loader", from: "", to: "fabric"}, true).detail).toBe("fabric")
  })
})

describe("changesTarget / entryIsEmpty", () => {
  it("detects minecraft and loader changes", () => {
    expect(changesTarget(entry({pack: [{field: "loader", from: "fabric", to: "neoforge"}]}))).toBe(true)
    expect(changesTarget(entry({pack: [{field: "loaderVersion", from: "1", to: "2"}]}))).toBe(false)
  })
  it("knows an empty entry", () => {
    expect(entryIsEmpty(entry({}))).toBe(true)
    expect(entryIsEmpty(entry({initial: true}))).toBe(false)
  })
})
