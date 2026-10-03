import "reflect-metadata"
import {describe, expect, it} from "vitest"
import {ref} from "vue"
import {Perm} from "@/lib/permissions.ts"
import {computeCan, usePackPermissions} from "./usePackPermissions.ts"

const pack = (...permissions: string[]) => ({permissions})

describe("computeCan", () => {
  it("allows only named permissions", () => {
    expect(computeCan(pack(Perm.PackView), Perm.PackView)).toBe(true)
    expect(computeCan(pack(Perm.PackView), Perm.PackModAdd)).toBe(false)
  })

  it("denies when permissions are missing", () => {
    expect(computeCan({permissions: undefined as unknown as string[]}, Perm.PackView)).toBe(false)
  })
})

describe("usePackPermissions", () => {
  it("reacts to permission changes", () => {
    const subject = ref(pack(Perm.PackView))
    const {can} = usePackPermissions(subject)
    expect(can(Perm.PackModAdd)).toBe(false)

    subject.value = pack(Perm.PackView, Perm.PackModAdd)
    expect(can(Perm.PackModAdd)).toBe(true)
  })
})
