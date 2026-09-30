import {plainToInstance} from "class-transformer";
import {apiClient} from "@/services/api.service";
import {
  OidcDiscoveryResult,
  OidcProvider,
  OidcPublicProvider,
  OrphanedUsers,
  UserIdentity,
} from "@/interfaces/oidc";
import type {OidcProviderRequest} from "@/interfaces/requests";

const PROVIDERS = 'v1/admin/oidc/providers'

export async function fetchAuthConfig(): Promise<OidcPublicProvider[]> {
  const response = await apiClient.get('v1/auth/config')
  return plainToInstance(OidcPublicProvider, response.data as unknown[])
}

export async function listOidcProviders(): Promise<OidcProvider[]> {
  const response = await apiClient.get(PROVIDERS)
  return plainToInstance(OidcProvider, response.data as unknown[])
}

export async function createOidcProvider(request: OidcProviderRequest): Promise<OidcProvider> {
  const response = await apiClient.post(PROVIDERS, request)
  return plainToInstance(OidcProvider, response.data)
}

export async function updateOidcProvider(id: number, request: OidcProviderRequest): Promise<OidcProvider> {
  const response = await apiClient.put(`${PROVIDERS}/${id}`, request)
  return plainToInstance(OidcProvider, response.data)
}

export async function deleteOidcProvider(id: number): Promise<void> {
  await apiClient.delete(`${PROVIDERS}/${id}`)
}

export async function fetchOrphanedUserCount(id: number): Promise<number> {
  const response = await apiClient.get(`${PROVIDERS}/${id}/orphaned-users`)
  return plainToInstance(OrphanedUsers, response.data).count
}

export async function testOidcDiscovery(issuerUrl: string): Promise<OidcDiscoveryResult> {
  const response = await apiClient.post(`${PROVIDERS}/test`, {issuerUrl})
  return plainToInstance(OidcDiscoveryResult, response.data)
}

export async function fetchUserIdentities(userId: number): Promise<UserIdentity[]> {
  const response = await apiClient.get(`v1/admin/users/${userId}/identities`)
  return plainToInstance(UserIdentity, response.data as unknown[])
}

export async function unlinkUserIdentity(userId: number, identityId: number): Promise<void> {
  await apiClient.delete(`v1/admin/users/${userId}/identities/${identityId}`)
}
