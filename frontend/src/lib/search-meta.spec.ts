import {describe, expect, it} from "vitest"
import {formatAuthorLine, formatDownloads, isPackLoader, limitChips} from "./search-meta.ts"

describe("formatDownloads", () => {
  it.each([
    [undefined, ""],
    [0, ""],
    [-4, ""],
    [NaN, ""],
    [1, "1 download"],
    [999, "999 downloads"],
    [1200, "1.2K downloads"],
    [1234567, "1.2M downloads"],
    [2_500_000_000, "2.5B downloads"],
  ])("%s -> %s", (input, expected) => {
    expect(formatDownloads(input)).toBe(expected)
  })
})

describe("formatAuthorLine", () => {
  it("joins author and downloads", () => {
    expect(formatAuthorLine("alice", 1200000)).toBe("by alice · 1.2M downloads")
  })
  it("handles missing parts", () => {
    expect(formatAuthorLine("alice")).toBe("by alice")
    expect(formatAuthorLine(undefined, 5000)).toBe("5K downloads")
    expect(formatAuthorLine()).toBe("")
  })
})

describe("isPackLoader", () => {
  it("matches case-insensitively", () => {
    expect(isPackLoader("Fabric", "fabric")).toBe(true)
  })
  it("rejects mismatch or missing pack loader", () => {
    expect(isPackLoader("forge", "fabric")).toBe(false)
    expect(isPackLoader("forge", undefined)).toBe(false)
  })
})

describe("limitChips", () => {
  it("limits and counts hidden", () => {
    expect(limitChips(["a", "b", "c", "d"], 3)).toEqual({shown: ["a", "b", "c"], hidden: 1})
  })
  it("handles undefined and short lists", () => {
    expect(limitChips(undefined, 2)).toEqual({shown: [], hidden: 0})
    expect(limitChips(["a"], 2)).toEqual({shown: ["a"], hidden: 0})
  })
})
