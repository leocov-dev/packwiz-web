import {describe, expect, it} from "vitest"
import {
  canRebaseSnapshot,
  canRevertSnapshot,
  rebaseConfirmText,
  fieldLabel,
  formatFieldValue,
  revertConfirmText,
  revertLabel,
  snapshotReasonIcon,
  snapshotReasonLabel,
  snapshotSubject,
  summarizeSnapshot,
} from "@/lib/snapshots.ts"
import type {SnapshotSummary} from "@/interfaces/snapshot.ts"

const summary = (over: Partial<SnapshotSummary>): SnapshotSummary => ({
  added: 0, removed: 0, changed: 0, packFields: [], ...over,
})

describe("snapshotReasonLabel / snapshotReasonIcon", () => {
  it("labels every known reason", () => {
    const reasons = [
      "baseline", "publish", "pack_edit", "mod_add", "mod_remove", "mod_update", "mod_side",
      "mod_option", "mod_pin", "rehash", "update_all", "migrate", "migrate_mods",
    ]
    for (const reason of reasons) {
      expect(snapshotReasonLabel(reason)).not.toBe(reason)
      expect(snapshotReasonIcon(reason)).not.toBe("mdi-history")
    }
  })

  it("falls back for an unknown reason", () => {
    expect(snapshotReasonLabel("something_new")).toBe("something_new")
    expect(snapshotReasonIcon("something_new")).toBe("mdi-history")
  })
})

describe("fieldLabel", () => {
  it("uses separate labels for the pack and mod version", () => {
    expect(fieldLabel("version", "pack")).toBe("Pack version")
    expect(fieldLabel("version", "mod")).toBe("Installed version")
  })

  it("labels nested mod fields and falls back to the raw name", () => {
    expect(fieldLabel("download.url", "mod")).toBe("Download URL")
    expect(fieldLabel("mcVersion", "pack")).toBe("Minecraft version")
    expect(fieldLabel("unknownField", "mod")).toBe("unknownField")
  })
})

describe("snapshotSubject", () => {
  it("prefers the mod name, then the slug", () => {
    expect(snapshotSubject({name: "Sodium", slug: "sodium"})).toBe("Sodium")
    expect(snapshotSubject({slug: "sodium"})).toBe("sodium")
    expect(snapshotSubject({name: ""})).toBe("")
  })

  it("is empty without a usable detail", () => {
    expect(snapshotSubject({})).toBe("")
    expect(snapshotSubject(null)).toBe("")
    expect(snapshotSubject(undefined)).toBe("")
    expect(snapshotSubject({name: 4, slug: 5})).toBe("")
  })
})

describe("summarizeSnapshot", () => {
  it("lists only non-zero counts then pack fields", () => {
    expect(summarizeSnapshot(summary({added: 2, changed: 1, packFields: ["mcVersion", "loader"]})))
      .toBe("2 added · 1 changed · Minecraft version · Loader")
  })

  it("reports removals", () => {
    expect(summarizeSnapshot(summary({removed: 3}))).toBe("3 removed")
  })

  it("says so when nothing changed", () => {
    expect(summarizeSnapshot(summary({}))).toBe("No changes")
  })

  it("tolerates a missing packFields list", () => {
    const s = {added: 1, removed: 0, changed: 0} as unknown as SnapshotSummary
    expect(summarizeSnapshot(s)).toBe("1 added")
  })
})

describe("formatFieldValue", () => {
  it.each([
    [null, "—"],
    [undefined, "—"],
    ["", "—"],
    [true, "Yes"],
    [false, "No"],
    [[], "—"],
    [["1.21", "1.20"], "1.21, 1.20"],
    ["fabric", "fabric"],
    [0, "0"],
  ])("formats %j as %s", (value, expected) => {
    expect(formatFieldValue(value)).toBe(expected)
  })
})

describe("revert rules", () => {
  it("never allows reverting an abandoned snapshot", () => {
    expect(canRevertSnapshot({isAbandoned: true}, true)).toBe(false)
  })

  it("needs the revert permission", () => {
    expect(canRevertSnapshot({isAbandoned: false}, false)).toBe(false)
    expect(canRevertSnapshot({isAbandoned: false}, true)).toBe(true)
  })

  it("words reverting to the head differently", () => {
    expect(revertLabel({isHead: true})).toBe("Discard unrecorded changes")
    expect(revertLabel({isHead: false})).toBe("Revert to this snapshot")
  })

  it("warns that a revert abandons later snapshots", () => {
    expect(revertConfirmText({isHead: false, seq: 3})).toContain("abandoned")
    expect(revertConfirmText({isHead: false, seq: 3})).toContain("#3")
    expect(revertConfirmText({isHead: true, seq: 3})).not.toContain("abandoned")
  })
})

describe("rebase", () => {
  it("needs manage permission and a live snapshot", () => {
    expect(canRebaseSnapshot({isAbandoned: true}, true)).toBe(false)
    expect(canRebaseSnapshot({isAbandoned: false}, false)).toBe(false)
    expect(canRebaseSnapshot({isAbandoned: false}, true)).toBe(true)
  })

  it("warns that earlier history is deleted", () => {
    expect(rebaseConfirmText({seq: 4})).toContain("#4")
    expect(rebaseConfirmText({seq: 4})).toContain("permanently deleted")
  })
})
