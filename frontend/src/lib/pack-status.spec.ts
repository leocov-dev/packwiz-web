import {describe, expect, it} from "vitest";
import {packStatusKey} from "@/lib/pack-status.ts";

describe("packStatusKey", () => {
  it("shows draft even when public", () => {
    expect(packStatusKey({status: "draft", isArchived: false, isPublic: true})).toBe("draft")
  })
  it("shows public for published public packs", () => {
    expect(packStatusKey({status: "published", isArchived: false, isPublic: true})).toBe("public")
  })
  it("shows published for private published packs", () => {
    expect(packStatusKey({status: "published", isArchived: false, isPublic: false})).toBe("published")
  })
  it("archived wins", () => {
    expect(packStatusKey({status: "published", isArchived: true, isPublic: true})).toBe("archived")
  })
})
