import {describe, expect, it} from "vitest"
import {checkTargetChange} from "./target-guard.ts"

const published = {status: "published", mcVersion: "1.21.1", loader: "fabric"}

describe("checkTargetChange", () => {
  it("allows anything on a draft", () => {
    expect(checkTargetChange({...published, status: "draft"}, "1.20.1", "forge"))
      .toEqual({blocked: null, confirm: false})
  })

  it("blocks a loader change", () => {
    expect(checkTargetChange(published, "1.21.1", "neoforge").blocked).toContain("loader")
  })

  it("ignores loader case", () => {
    expect(checkTargetChange(published, "1.21.1", "Fabric")).toEqual({blocked: null, confirm: false})
  })

  it("blocks a downgrade", () => {
    expect(checkTargetChange(published, "1.20.1", "fabric").blocked).toContain("older Minecraft")
  })

  it("asks to confirm an upgrade", () => {
    expect(checkTargetChange(published, "26.1", "fabric")).toEqual({blocked: null, confirm: true})
  })

  it("treats an uncomparable version as a change to confirm", () => {
    expect(checkTargetChange(published, "__latest__", "fabric")).toEqual({blocked: null, confirm: true})
  })
})
