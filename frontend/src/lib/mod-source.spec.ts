import {describe, expect, it} from "vitest"
import {buildRequest, isSearchResultInstalled, parseUrl} from "./mod-source.ts"

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

describe("isSearchResultInstalled", () => {
  it("returns true when result.installed is true", () => {
    expect(isSearchResultInstalled({slug: "fabric-api", projectId: "P7dR8mSH", installed: true})).toBe(true)
  })

  it("returns true when slug matches installed mods case-insensitively", () => {
    const installedMods = [{slug: "Fabric-Api", update: {}}]
    expect(isSearchResultInstalled({slug: "fabric-api", projectId: "P7dR8mSH"}, installedMods)).toBe(true)
  })

  it("returns true when projectId matches update['mod-id'] or update.modrinth['mod-id']", () => {
    const installedMods1 = [{slug: "custom-slug", update: {"mod-id": "P7dR8mSH"}}]
    expect(isSearchResultInstalled({slug: "fabric-api", projectId: "P7dR8mSH"}, installedMods1)).toBe(true)

    const installedMods2 = [{slug: "custom-slug", update: {modrinth: {"mod-id": "P7dR8mSH"}}}]
    expect(isSearchResultInstalled({slug: "fabric-api", projectId: "P7dR8mSH"}, installedMods2)).toBe(true)
  })

  it("returns false when mod is not installed", () => {
    const installedMods = [{slug: "sodium", update: {modrinth: {"mod-id": "AANobbMI"}}}]
    expect(isSearchResultInstalled({slug: "fabric-api", projectId: "P7dR8mSH"}, installedMods)).toBe(false)
  })

  it("returns false when installedMods list is empty or undefined", () => {
    expect(isSearchResultInstalled({slug: "fabric-api", projectId: "P7dR8mSH"}, [])).toBe(false)
    expect(isSearchResultInstalled({slug: "fabric-api", projectId: "P7dR8mSH"}, undefined)).toBe(false)
  })
})
