import {describe, expect, it} from "vitest";
import type {PackStatus} from "@/interfaces/pack.ts";
import {packStatusKey} from "@/lib/pack-status.ts";

describe("packStatusKey", () => {
  it("shows draft even when public", () => {
    expect(packStatusKey({status: "draft" as PackStatus, isArchived: false, isPublic: true})).toBe("draft")
  })
  it("shows public for published public packs", () => {
    expect(packStatusKey({status: "published" as PackStatus, isArchived: false, isPublic: true})).toBe("public")
  })
  it("shows published for private published packs", () => {
    expect(packStatusKey({status: "published" as PackStatus, isArchived: false, isPublic: false})).toBe("published")
  })
  it("archived wins", () => {
    expect(packStatusKey({status: "published" as PackStatus, isArchived: true, isPublic: true})).toBe("archived")
  })
})
