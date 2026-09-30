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
}))

import {
  createOidcProvider, deleteOidcProvider, fetchAuthConfig, fetchOrphanedUserCount,
  fetchUserIdentities, listOidcProviders, testOidcDiscovery, unlinkUserIdentity, updateOidcProvider,
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
