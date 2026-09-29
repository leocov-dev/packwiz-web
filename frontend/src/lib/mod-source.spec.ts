import {describe, expect, it} from "vitest"
import {addedKey, buildRequest, getResultState, normalizeSearchQuery, isSearchResultInstalled, parseUrl, searchEmptyState} from "./mod-source.ts"

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

  it("ignores a recognized domain in the query string or path", () => {
    expect(parseUrl("https://example.com/?u=modrinth.com")).toBe("")
    expect(parseUrl("https://example.com/github.com/owner/repo")).toBe("")
  })

  it("ignores lookalike hosts", () => {
    expect(parseUrl("https://notmodrinth.com/mod/x")).toBe("")
    expect(parseUrl("https://modrinth.com.evil.io/mod/x")).toBe("")
  })

  it("returns empty string for an invalid URL", () => {
    expect(parseUrl("modrinth.com/mod/sodium")).toBe("")
    expect(parseUrl("not a url")).toBe("")
  })

  it("accepts www and other subdomains", () => {
    expect(parseUrl("https://www.modrinth.com/mod/sodium")).toBe("Modrinth")
    expect(parseUrl("https://legacy.curseforge.com/minecraft/mc-mods/jei")).toBe("Curseforge")
    expect(parseUrl("https://www.github.com/owner/repo")).toBe("Github")
  })

  it("accepts http and https but not other protocols", () => {
    expect(parseUrl("http://modrinth.com/mod/sodium")).toBe("Modrinth")
    expect(parseUrl("ftp://modrinth.com/mod/sodium")).toBe("")
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

  it("returns true when projectId matches update['project-id'] or update.curseforge['project-id']", () => {
    const installedMods1 = [{slug: "custom-slug", update: {"project-id": "238222"}}]
    expect(isSearchResultInstalled({slug: "jei", projectId: "238222"}, installedMods1)).toBe(true)

    const installedMods2 = [{slug: "custom-slug", update: {curseforge: {"project-id": 238222}}}]
    expect(isSearchResultInstalled({slug: "jei", projectId: "238222"}, installedMods2)).toBe(true)
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

describe("getResultState", () => {
  const installed = [{slug: "sodium"}]

  it("is available for an unknown result", () => {
    expect(getResultState({slug: "iris", projectId: "x"}, installed, [], "modrinth")).toBe("available")
  })

  it("is installed when present in pack mods", () => {
    expect(getResultState({slug: "Sodium"}, installed, [], "modrinth")).toBe("installed")
  })

  it("is added when added locally, even if pack mods are stale", () => {
    expect(getResultState({slug: "iris"}, installed, [addedKey("modrinth", "iris")], "modrinth")).toBe("added")
  })

  it("does not mark a same-slug result from another source as added", () => {
    expect(getResultState({slug: "iris"}, installed, [addedKey("modrinth", "iris")], "curseforge")).toBe("available")
  })

  it("prefers added when slug is both installed and added", () => {
    expect(getResultState({slug: "sodium"}, installed, [addedKey("modrinth", "sodium")], "modrinth")).toBe("added")
  })

  it("handles undefined installed mods", () => {
    expect(getResultState({slug: "iris"}, undefined, [], "modrinth")).toBe("available")
  })
})

describe("searchEmptyState", () => {
  it("prompts for a longer query", () => {
    expect(searchEmptyState("a", false, 0, false)).toBe("short-query")
    expect(searchEmptyState("", false, 0, true)).toBe("short-query")
    expect(searchEmptyState("   ", false, 0, true)).toBe("short-query")
  })

  it("reports no results after a completed search", () => {
    expect(searchEmptyState("sodium", false, 0, true)).toBe("no-results")
  })

  it("shows nothing while loading, before search, or with results", () => {
    expect(searchEmptyState("sodium", true, 0, true)).toBe("none")
    expect(searchEmptyState("sodium", false, 0, false)).toBe("none")
    expect(searchEmptyState("sodium", false, 3, true)).toBe("none")
  })
})

describe("normalizeSearchQuery", () => {
  it("trims whitespace", () => {
    expect(normalizeSearchQuery("  sodium ")).toBe("sodium")
  })

  it("reduces a whitespace-only query to empty", () => {
    expect(normalizeSearchQuery("   ")).toBe("")
  })

  it("handles null and undefined", () => {
    expect(normalizeSearchQuery(null)).toBe("")
    expect(normalizeSearchQuery(undefined)).toBe("")
  })
})
