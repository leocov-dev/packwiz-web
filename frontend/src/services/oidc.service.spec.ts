import "reflect-metadata"
import {beforeEach, describe, expect, it, vi} from "vitest"

const {get, post, put, del} = vi.hoisted(() => ({
  get: vi.fn(),
  post: vi.fn(),
  put: vi.fn(),
  del: vi.fn(),
}))

vi.mock("@/services/api.service", () => ({
  apiClient: {get, post, put, delete: del},
  baseUrl: 'https://pw.example.com',
}))

import {
  createOidcProvider, deleteOidcProvider, fetchAuthConfig, fetchOrphanedUserCount,
  fetchMyIdentities, fetchUserIdentities, listOidcProviders, oidcLoginUrl, startLinkIdentity, unlinkMyIdentity, testOidcDiscovery, unlinkUserIdentity, updateOidcProvider,
} from "@/services/oidc.service.ts"
import {OidcProvider, OidcPublicProvider, UserIdentity} from "@/interfaces/oidc.ts"
import type {OidcProviderRequest} from "@/interfaces/requests.ts"

const request: OidcProviderRequest = {
  slug: 'kc', displayName: 'Keycloak', issuerUrl: 'https://idp.example.com', clientId: 'cid',
  clientSecret: '', scopes: 'openid', enabled: true, autoCreateUsers: false, linkByEmail: false,
}

beforeEach(() => vi.resetAllMocks())

describe("public config", () => {
  it("hydrates providers from the unauthenticated endpoint", async () => {
    get.mockResolvedValue({data: [{slug: 'kc', displayName: 'Keycloak'}]})
    const res = await fetchAuthConfig()
    expect(get).toHaveBeenCalledWith('v1/auth/config')
    expect(res[0]).toBeInstanceOf(OidcPublicProvider)
  })
})

describe("admin provider calls", () => {
  it("lists providers", async () => {
    get.mockResolvedValue({data: [{id: 1, slug: 'kc'}]})
    const res = await listOidcProviders()
    expect(get).toHaveBeenCalledWith('v1/admin/oidc/providers')
    expect(res[0]).toBeInstanceOf(OidcProvider)
  })

  it("creates and updates with the request body", async () => {
    post.mockResolvedValue({data: {id: 2}})
    put.mockResolvedValue({data: {id: 2}})
    await createOidcProvider(request)
    await updateOidcProvider(2, request)
    expect(post).toHaveBeenCalledWith('v1/admin/oidc/providers', request)
    expect(put).toHaveBeenCalledWith('v1/admin/oidc/providers/2', request)
  })

  it("deletes, counts orphaned users and tests discovery", async () => {
    del.mockResolvedValue({})
    get.mockResolvedValue({data: {count: 4}})
    post.mockResolvedValue({data: {issuer: 'x', authorizationEndpoint: 'a', tokenEndpoint: 't'}})

    await deleteOidcProvider(3)
    expect(del).toHaveBeenCalledWith('v1/admin/oidc/providers/3')

    expect(await fetchOrphanedUserCount(3)).toBe(4)
    expect(get).toHaveBeenCalledWith('v1/admin/oidc/providers/3/orphaned-users')

    const res = await testOidcDiscovery('https://idp.example.com')
    expect(post).toHaveBeenCalledWith('v1/admin/oidc/providers/test', {issuerUrl: 'https://idp.example.com'})
    expect(res.tokenEndpoint).toBe('t')
  })
})

describe("admin user identities", () => {
  it("lists and unlinks", async () => {
    get.mockResolvedValue({data: [{id: 9, providerName: 'Keycloak'}]})
    del.mockResolvedValue({})

    const res = await fetchUserIdentities(5)
    expect(get).toHaveBeenCalledWith('v1/admin/users/5/identities')
    expect(res[0]).toBeInstanceOf(UserIdentity)

    await unlinkUserIdentity(5, 9)
    expect(del).toHaveBeenCalledWith('v1/admin/users/5/identities/9')
  })
})

describe("self-service identities", () => {
  it("builds a login url with an optional redirect", () => {
    expect(oidcLoginUrl('kc')).toBe('https://pw.example.com/api/v1/auth/oidc/kc/login')
    expect(oidcLoginUrl('kc', '/packs/1?x=y')).toMatch(/\/login\?redirect=%2Fpacks%2F1%3Fx%3Dy$/)
    expect(oidcLoginUrl('a/b')).toContain('/oidc/a%2Fb/login')
  })

  it("lists, starts linking and unlinks", async () => {
    get.mockResolvedValue({data: [{id: 1, providerSlug: 'kc'}]})
    post.mockResolvedValue({data: {redirectUrl: 'https://idp.example.com/auth?x=1'}})
    del.mockResolvedValue({})

    const res = await fetchMyIdentities()
    expect(get).toHaveBeenCalledWith('v1/user/identities')
    expect(res[0]).toBeInstanceOf(UserIdentity)

    expect(await startLinkIdentity('kc')).toBe('https://idp.example.com/auth?x=1')
    expect(post).toHaveBeenCalledWith('v1/user/identities/kc/link')

    await unlinkMyIdentity(1)
    expect(del).toHaveBeenCalledWith('v1/user/identities/1')
  })
})
