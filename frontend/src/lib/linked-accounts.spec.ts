import {describe, expect, it} from "vitest"
import {linkableProviders} from "@/lib/linked-accounts.ts"
import type {OidcPublicProvider, UserIdentity} from "@/interfaces/oidc.ts"

const provider = (slug: string) => ({slug, displayName: slug}) as OidcPublicProvider
const identity = (providerSlug: string) => ({providerSlug}) as UserIdentity

describe("linkableProviders", () => {
  it("excludes providers that already have an identity", () => {
    const res = linkableProviders([provider("a"), provider("b")], [identity("a")])
    expect(res.map(p => p.slug)).toEqual(["b"])
  })

  it("offers everything when nothing is linked", () => {
    expect(linkableProviders([provider("a")], [])).toHaveLength(1)
  })

  it("offers nothing when there are no providers", () => {
    expect(linkableProviders([], [identity("a")])).toEqual([])
  })
})
