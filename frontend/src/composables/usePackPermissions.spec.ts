import "reflect-metadata"
import {describe, expect, it} from "vitest"
import {ref} from "vue"
import {PackPermission} from "@/interfaces/pack.ts"
import {computeCanEdit, computeHasEditAccess, usePackPermissions} from "./usePackPermissions.ts"

const pack = (currentUserPermission: PackPermission, isArchived = false) => ({currentUserPermission, isArchived})

describe("computeCanEdit", () => {
  it("denies static and view permission", () => {
    expect(computeCanEdit(pack(PackPermission.STATIC))).toBe(false)
    expect(computeCanEdit(pack(PackPermission.VIEW))).toBe(false)
  })

  it("allows edit permission", () => {
    expect(computeCanEdit(pack(PackPermission.EDIT))).toBe(true)
  })

  it("denies archived packs", () => {
    expect(computeCanEdit(pack(PackPermission.EDIT, true))).toBe(false)
    expect(computeCanEdit(pack(PackPermission.VIEW, true))).toBe(false)
  })
})

describe("computeHasEditAccess", () => {
  it("ignores archived state", () => {
    expect(computeHasEditAccess(pack(PackPermission.EDIT, true))).toBe(true)
    expect(computeHasEditAccess(pack(PackPermission.VIEW, true))).toBe(false)
  })
})

describe("usePackPermissions", () => {
  it("reacts to permission and archived changes", () => {
    const subject = ref(pack(PackPermission.VIEW))
    const {canEdit, hasEditAccess} = usePackPermissions(subject)
    expect(canEdit.value).toBe(false)
    expect(hasEditAccess.value).toBe(false)

    subject.value = pack(PackPermission.EDIT)
    expect(canEdit.value).toBe(true)
    expect(hasEditAccess.value).toBe(true)

    subject.value = pack(PackPermission.EDIT, true)
    expect(canEdit.value).toBe(false)
    expect(hasEditAccess.value).toBe(true)
  })
})
