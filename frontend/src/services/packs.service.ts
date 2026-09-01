import {
  AllPacksResponse,
  MigrateDryRunResponse,
  MigrateJobStatusResponse,
  MigrateResponse,
  Pack,
  PackCollaboratorsResponse,
  PackPermission,
  PackResponse,
  UserSearchResponse
} from "@/interfaces/pack";
import {apiClient} from "@/services/api.service";
import {plainToInstance} from "class-transformer";
import type {EditPackRequest, MigratePackRequest, NewPackRequest, RehashRequest} from "@/interfaces/requests.ts";


export async function fetchAllPacks(
  statusList: string[],
  archived: boolean = false,
  search: string = '',
): Promise<AllPacksResponse> {

  let url = 'v1/packwiz/pack'

  const params = new URLSearchParams();
  if (statusList.length > 0) {
    statusList.forEach(status => params.append('status', status));
  }
  if (archived) {
    params.append('archived', 'true');
  }
  if (search !== "") {
    params.append('search', search);
  }

  if (params.size > 0) {
    url += `?${params.toString()}`
  }

  const response = await apiClient.get(url);
  return plainToInstance(AllPacksResponse, response.data)

}

export async function fetchOnePack(packId: number, skipMods: boolean = false): Promise<PackResponse> {
  let url = `v1/packwiz/pack/${packId}`

  const params = new URLSearchParams();
  if (skipMods) {
    params.append('skipMods', 'true');
  }

  if (params.size > 0) {
    url += `?${params.toString()}`
  }

  const response = await apiClient.get(url);
  return plainToInstance(PackResponse, response.data)
}


export async function getPackPublicLink(packId: number): Promise<string> {
  const response = await apiClient.get(`v1/packwiz/pack/${packId}/link`);
  return response.data['link']
}


export async function linkToClipboard(packId: number) {
  const link = await getPackPublicLink(packId)
  await navigator.clipboard.writeText(link)
}

export async function openPublicLink(packId: number) {
  const link = await getPackPublicLink(packId)
  window.open(link, '_blank')
}

export async function newPack(request: NewPackRequest): Promise<Pack> {
  const response = await apiClient.post('v1/packwiz/pack', request)
  return plainToInstance(Pack, response.data)
}

export async function editPack(packId: number, request: EditPackRequest) {
  return apiClient.patch(`v1/packwiz/pack/${packId}/edit`, request)
}

export async function archivePack(packId: number) {
  return apiClient.delete(`v1/packwiz/pack/${packId}`)
}

export async function unArchivePack(packId: number) {
  return apiClient.patch(`v1/packwiz/pack/${packId}/unarchive`)
}

export async function publishPack(packId: number) {
  return apiClient.patch(`v1/packwiz/pack/${packId}/publish`)
}

export async function convertPackToDraft(packId: number) {
  return apiClient.patch(`v1/packwiz/pack/${packId}/draft`)
}

export async function makePackPublic(packId: number) {
  return apiClient.patch(`v1/packwiz/pack/${packId}/public`)
}

export async function makePackPrivate(packId: number) {
  return apiClient.patch(`v1/packwiz/pack/${packId}/private`)
}

export async function updateAllMods(packId: number) {
  return apiClient.patch(`v1/packwiz/pack/${packId}/update-all`)
}

export async function rehashAllMods(packId: number, request: RehashRequest) {
  return apiClient.patch(`v1/packwiz/pack/${packId}/rehash`, request)
}

export async function migratePack(packId: number, request: MigratePackRequest): Promise<MigrateResponse> {
  const response = await apiClient.patch(`v1/packwiz/pack/${packId}/migrate`, request)
  return plainToInstance(MigrateResponse, response.data)
}

export async function migrateDryRun(packId: number, request: MigratePackRequest): Promise<MigrateDryRunResponse> {
  const response = await apiClient.post(`v1/packwiz/pack/${packId}/migrate/dry-run`, request)
  return plainToInstance(MigrateDryRunResponse, response.data)
}

export async function getMigrateJobStatus(packId: number, jobId: number): Promise<MigrateJobStatusResponse> {
  const response = await apiClient.get(`v1/packwiz/pack/${packId}/migrate/job/${jobId}`)
  return plainToInstance(MigrateJobStatusResponse, response.data)
}

export async function getPackCollaborators(packId: number): Promise<PackCollaboratorsResponse> {
  const response = await apiClient.get(`v1/packwiz/pack/${packId}/users`);
  return plainToInstance(PackCollaboratorsResponse, response.data)
}

export async function searchUsersForPack(packId: number, query: string): Promise<UserSearchResponse> {
  const params = new URLSearchParams({q: query});
  const response = await apiClient.get(`v1/packwiz/pack/${packId}/users/search?${params.toString()}`);
  return plainToInstance(UserSearchResponse, response.data)
}

export async function addPackCollaborator(packId: number, userId: number, permission: PackPermission) {
  return apiClient.post(`v1/packwiz/pack/${packId}/users`, {userId, permission})
}

export async function updateCollaboratorPermission(packId: number, userId: number, permission: PackPermission) {
  return apiClient.patch(`v1/packwiz/pack/${packId}/users/${userId}`, {permission})
}

export async function removeCollaborator(packId: number, userId: number) {
  return apiClient.delete(`v1/packwiz/pack/${packId}/users/${userId}`)
}
