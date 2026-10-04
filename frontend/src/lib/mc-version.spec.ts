import {describe, expect, it} from "vitest"
import {compareMinecraftVersions} from "./mc-version.ts"

describe("compareMinecraftVersions", () => {
  it.each([
    ["1.21.1", "1.21.1", 0],
    ["1.21", "1.21.0", 0],
    ["1.20.4", "1.20.5", -1],
    ["1.21.10", "1.21.9", 1],
    ["1.21.1", "26.1", -1],
    ["26.1", "26.1.2", -1],
    ["26.3-rc-3", "26.3", -1],
    ["26.3-pre-1", "26.3-rc-1", -1],
    ["26.3-snapshot-10", "26.3-snapshot-9", 1],
    ["1.21-pre1", "1.21", -1],
  ])("%s vs %s", (a, b, want) => {
    expect(compareMinecraftVersions(a, b)).toBe(want)
  })

  it.each([["24w14a", "1.21"], ["1.21", ""], ["26.1-beta-1", "26.1"]])("can't compare %s / %s", (a, b) => {
    expect(compareMinecraftVersions(a, b)).toBeNull()
  })
})
