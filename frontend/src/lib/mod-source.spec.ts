import {describe, expect, it} from "vitest"
import {buildRequest, parseUrl} from "./mod-source.ts"

describe("parseUrl", () => {
  it("recognizes a curseforge URL", () => {
    expect(parseUrl("https://www.curseforge.com/minecraft/mc-mods/jei")).toBe("Curseforge")
  })

  it("recognizes a modrinth URL", () => {
    expect(parseUrl("https://modrinth.com/mod/sodium")).toBe("Modrinth")
  })

  it("recognizes a github URL", () => {
    expect(parseUrl("https://github.com/owner/repo")).toBe("Github")
  })

  it("returns empty string for an unrecognized URL", () => {
    expect(parseUrl("https://example.com/mod")).toBe("")
  })

  it("returns empty string for an empty URL", () => {
    expect(parseUrl("")).toBe("")
  })
})

describe("buildRequest", () => {
  it("builds a curseforge request", () => {
    expect(buildRequest("Curseforge", "https://www.curseforge.com/minecraft/mc-mods/jei"))
      .toEqual({request: {curseforge: {url: "https://www.curseforge.com/minecraft/mc-mods/jei"}}})
  })

  it("builds a modrinth request", () => {
    expect(buildRequest("Modrinth", "https://modrinth.com/mod/sodium"))
      .toEqual({request: {modrinth: {url: "https://modrinth.com/mod/sodium"}}})
  })

  it("builds a github request", () => {
    expect(buildRequest("Github", "https://github.com/owner/repo"))
      .toEqual({request: {github: {url: "https://github.com/owner/repo"}}})
  })

  it("returns an error for an empty/invalid source", () => {
    expect(buildRequest("", "https://example.com/mod"))
      .toEqual({error: "Invalid mod source: "})
  })
})
