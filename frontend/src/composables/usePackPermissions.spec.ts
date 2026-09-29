import "reflect-metadata"
import {describe, expect, it, vi} from "vitest"
import {PackPermission} from "@/interfaces/pack.ts"
vi.mock("@/stores/auth.ts", () => ({useAuthStore: () => ({user: null})}))

import {computeCanEdit, computeHasEditAccess} from "./usePackPermissions.ts"

const pack = (currentUserPermission: PackPermission, isArchived = false) => ({currentUserPermission, isArchived})

describe("computeCanEdit", () => {
  it("denies static and view permission", () => {
    expect(computeCanEdit(pack(PackPermission.STATIC), false)).toBe(false)
    expect(computeCanEdit(pack(PackPermission.VIEW), false)).toBe(false)
  })

  it("allows edit permission", () => {
    expect(computeCanEdit(pack(PackPermission.EDIT), false)).toBe(true)
  })

  it("allows admin without collaborator permission", () => {
    expect(computeCanEdit(pack(PackPermission.STATIC), true)).toBe(true)
    expect(computeCanEdit(pack(PackPermission.VIEW), true)).toBe(true)
  })

  it("denies everyone on archived packs", () => {
    expect(computeCanEdit(pack(PackPermission.EDIT, true), false)).toBe(false)
    expect(computeCanEdit(pack(PackPermission.EDIT, true), true)).toBe(false)
    expect(computeCanEdit(pack(PackPermission.VIEW, true), true)).toBe(false)
  })
})

describe("computeHasEditAccess", () => {
  it("ignores archived state", () => {
    expect(computeHasEditAccess(pack(PackPermission.EDIT, true), false)).toBe(true)
    expect(computeHasEditAccess(pack(PackPermission.VIEW, true), false)).toBe(false)
    expect(computeHasEditAccess(pack(PackPermission.VIEW, true), true)).toBe(true)
  })
})
