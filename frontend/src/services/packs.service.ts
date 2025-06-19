import {AllPacksResponse, PackResponse} from "@/interfaces/pack";
import {apiClient} from "@/services/api.service";
import {plainToInstance} from "class-transformer";
import type {EditPackRequest, NewPackRequest} from "@/interfaces/requests.ts";


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

export async function newPack(request: NewPackRequest) {
  return apiClient.post('v1/packwiz/pack', request)
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
