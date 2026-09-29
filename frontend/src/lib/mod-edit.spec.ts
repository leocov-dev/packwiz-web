import {describe, expect, it} from "vitest"
import {
  buildOptionRequest,
  describeSaveFailure,
  describeUpdateResult,
  diffModEdit,
  editValuesFromMod,
  MAX_OPTION_DESCRIPTION,
  modWasUpdated,
  runSaveSteps,
  type ModEditValues,
} from "@/lib/mod-edit.ts"
import type {Mod} from "@/interfaces/pack.ts"

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

describe("buildOptionRequest boundary", () => {
  it("passes a 500-char description through unchanged", () => {
    const description = "a".repeat(MAX_OPTION_DESCRIPTION)
    expect(buildOptionRequest({...base, optional: true, description}).description).toHaveLength(500)
  })
})

describe("editValuesFromMod", () => {
  it("handles null option and undefined pinned", () => {
    const mod = {side: "server", pinned: undefined, option: null} as unknown as Mod
    expect(editValuesFromMod(mod)).toEqual({
      side: "server", pinned: false, optional: false, description: "", default: false,
    })
  })
  it("maps option fields", () => {
    const mod = {side: "both", pinned: true, option: {optional: true, description: "d", default: true}} as unknown as Mod
    expect(editValuesFromMod(mod)).toEqual({
      side: "both", pinned: true, optional: true, description: "d", default: true,
    })
  })
})

describe("runSaveSteps", () => {
  const ok = (label: string, calls: string[]) => ({label, run: async () => { calls.push(label) }})
  const bad = (label: string, calls: string[]) => ({
    label,
    run: async () => { calls.push(label); throw new Error("boom") },
  })

  it("runs all steps when all succeed", async () => {
    const calls: string[] = []
    const r = await runSaveSteps([ok("Side change", calls), ok("Options", calls)])
    expect(r.every(x => x.ok)).toBe(true)
    expect(calls).toEqual(["Side change", "Options"])
  })

  it("still runs later steps when one fails", async () => {
    const calls: string[] = []
    const r = await runSaveSteps([ok("Side change", calls), bad("Pin change", calls), ok("Options", calls)])
    expect(calls).toEqual(["Side change", "Pin change", "Options"])
    expect(r.map(x => x.ok)).toEqual([true, false, true])
    expect(r[1].message).toBe("boom")
  })

  it("edit-then-retry resends only the failed step", async () => {
    let baseline = {...base}
    let pinFails = true
    const target = {...base, side: "client" as const, pinned: true}
    const build = (cur: ModEditValues) => {
      const d = diffModEdit(baseline, cur)
      const steps = []
      if (d.side) steps.push({label: "Side change", run: async () => { baseline = {...baseline, side: cur.side} }})
      if (d.pinned) steps.push({
        label: "Pin change",
        run: async () => {
          if (pinFails) throw new Error("nope")
          baseline = {...baseline, pinned: cur.pinned}
        },
      })
      return steps
    }
    const first = await runSaveSteps(build(target))
    expect(first.map(x => x.ok)).toEqual([true, false])
    pinFails = false
    const retry = build(target)
    expect(retry.map(s => s.label)).toEqual(["Pin change"])
    expect((await runSaveSteps(retry)).every(x => x.ok)).toBe(true)
    expect(diffModEdit(baseline, target).any).toBe(false)
  })

  it("uses custom message mapper", async () => {
    const r = await runSaveSteps([bad("Side change", [])], () => "mapped")
    expect(r[0].message).toBe("mapped")
  })
})

describe("describeSaveFailure", () => {
  it("reports each step in order", () => {
    expect(describeSaveFailure([
      {label: "Side change", ok: true},
      {label: "Pin change", ok: false, message: "boom"},
      {label: "Options", ok: true},
    ])).toBe("Side change saved. Pin change failed: boom. Options saved.")
  })
})

describe("update result", () => {
  const a = {fileName: "a.jar", update: {hash: "1", id: 2}}
  it("detects no change regardless of key order", () => {
    expect(modWasUpdated(a, {fileName: "a.jar", update: {id: 2, hash: "1"}})).toBe(false)
    expect(describeUpdateResult("Foo", a, a)).toBe("Foo is already up to date")
  })
  it("detects file change", () => {
    expect(describeUpdateResult("Foo", a, {...a, fileName: "b.jar"})).toBe("Updated Foo to b.jar")
  })
  it("detects update metadata change", () => {
    expect(modWasUpdated(a, {...a, update: {hash: "2", id: 2}})).toBe(true)
  })
})
