import {describe, expect, it} from "vitest"
import {
  DEFAULT_OIDC_SCOPES, deleteConfirmationText, normalizeScopes, providerStatus, slugify,
  validateIssuerUrl, validateScopes, validateSlug,
} from "@/lib/oidc-form.ts"

describe("slugify", () => {
  it.each([
    ["Keycloak", "keycloak"],
    ["My  Company SSO!", "my-company-sso"],
    ["  --Google--  ", "google"],
    ["", ""],
  ])("%s -> %s", (input, want) => expect(slugify(input)).toBe(want))

  it("never produces an invalid slug when truncating", () => {
    const slug = slugify("a".repeat(63) + " b")
    expect(slug.length).toBeLessThanOrEqual(64)
    expect(validateSlug(slug)).toBe(true)
  })
})

describe("validateSlug", () => {
  it.each(["a", "keycloak", "my-idp-2"])("accepts %s", (s) => expect(validateSlug(s)).toBe(true))
  it.each(["", "Upper", "-a", "a-", "a b", "a/b", "a".repeat(65)])("rejects %s", (s) =>
    expect(validateSlug(s)).not.toBe(true))
})

describe("validateIssuerUrl", () => {
  it.each([
    "https://idp.example.com",
    "https://idp.example.com/realms/main",
    "http://localhost:8080/realms/dev",
    "http://127.0.0.1:9000",
  ])("accepts %s", (u) => expect(validateIssuerUrl(u)).toBe(true))

  it.each([
    "",
    "idp.example.com",
    "http://idp.example.com",
    "ftp://idp.example.com",
    "https://idp.example.com?x=1",
    "https://idp.example.com#frag",
  ])("rejects %s", (u) => expect(validateIssuerUrl(u)).not.toBe(true))
})

describe("scopes", () => {
  it("defaults when blank and de-duplicates", () => {
    expect(normalizeScopes("  ")).toBe(DEFAULT_OIDC_SCOPES)
    expect(normalizeScopes("openid email  email")).toBe("openid email")
  })

  it("requires openid", () => {
    expect(validateScopes("")).toBe(true)
    expect(validateScopes("openid groups")).toBe(true)
    expect(validateScopes("profile email")).not.toBe(true)
  })
})

describe("providerStatus", () => {
  it("prioritises broken over enabled", () => {
    expect(providerStatus({enabled: true, broken: true})).toBe('broken')
    expect(providerStatus({enabled: true, broken: false})).toBe('enabled')
    expect(providerStatus({enabled: false, broken: false})).toBe('disabled')
  })
})

describe("deleteConfirmationText", () => {
  it("warns only when users would be stranded", () => {
    expect(deleteConfirmationText("Corp", 0)).not.toContain("locked out")
    expect(deleteConfirmationText("Corp", 1)).toContain("1 user has")
    expect(deleteConfirmationText("Corp", 3)).toContain("3 users have")
  })
})
