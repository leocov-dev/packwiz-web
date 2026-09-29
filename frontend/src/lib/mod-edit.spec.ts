import {describe, expect, it} from "vitest"
import {buildOptionRequest, describeSaveFailure, diffModEdit, type ModEditValues} from "@/lib/mod-edit.ts"

const base: ModEditValues = {side: "both", pinned: false, optional: false, description: "", default: false}

describe("diffModEdit", () => {
  it("reports nothing when equal", () => {
    expect(diffModEdit(base, {...base})).toEqual({side: false, pinned: false, option: false, any: false})
  })
  it("detects side and pin", () => {
    const d = diffModEdit(base, {...base, side: "client", pinned: true})
    expect(d).toMatchObject({side: true, pinned: true, option: false, any: true})
  })
  it("ignores hidden description/default when not optional", () => {
    expect(diffModEdit(base, {...base, description: "x", default: true}).any).toBe(false)
  })
  it("detects option fields when optional", () => {
    const opt = {...base, optional: true}
    expect(diffModEdit(opt, {...opt, description: "x"}).option).toBe(true)
    expect(diffModEdit(opt, {...opt, default: true}).option).toBe(true)
  })
  it("detects toggling optional off", () => {
    expect(diffModEdit({...base, optional: true, description: "x"}, base).option).toBe(true)
  })
})

describe("buildOptionRequest", () => {
  it("clears hidden fields when not optional", () => {
    expect(buildOptionRequest({...base, description: "x", default: true}))
      .toEqual({optional: false, description: "", default: false})
  })
  it("keeps fields when optional", () => {
    expect(buildOptionRequest({...base, optional: true, description: "x", default: true}))
      .toEqual({optional: true, description: "x", default: true})
  })
})

describe("describeSaveFailure", () => {
  it("lists saved, failed and remaining", () => {
    expect(describeSaveFailure(["Side"], "Pin change", "boom", ["Options"]))
      .toBe("Side saved. Pin change failed: boom. Not saved: Options.")
  })
  it("omits empty parts", () => {
    expect(describeSaveFailure([], "Side change", "boom")).toBe("Side change failed: boom.")
  })
})
