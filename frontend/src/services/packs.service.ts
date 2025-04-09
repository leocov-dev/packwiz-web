import {Packs, Pack} from "@/interfaces/pack";
import {apiClient} from "@/services/api.service";
import {plainToInstance} from "class-transformer";
import type {EditPackRequest, NewPackRequest} from "@/interfaces/requests.ts";


export async function fetchAllPacks(
  statusList: string[],
  archived: boolean = false,
  search: string = '',
): Promise<Packs> {

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
  return plainToInstance(Packs, response.data)

}

export async function fetchOnePack(slug: string): Promise<Pack> {
  const response = await apiClient.get(`v1/packwiz/pack/${slug}`);
  return plainToInstance(Pack, response.data)
}


export async function getPackPublicLink(slug: string): Promise<string> {
  const response = await apiClient.get(`v1/packwiz/pack/${slug}/link`);
  return response.data['link']
}


export async function linkToClipboard(slug: string) {
  const link = await getPackPublicLink(slug)
  await navigator.clipboard.writeText(link)
}

export async function openPublicLink(slug: string) {
  const link = await getPackPublicLink(slug)
  window.open(link, '_blank')
}

export async function newPack(request: NewPackRequest) {
  return apiClient.post('v1/packwiz/pack', request)
}

export async function editPack(slug: string, request: EditPackRequest) {
  return apiClient.patch(`v1/packwiz/pack/${slug}/edit`, request)
}

export async function archivePack(slug: string) {
  return apiClient.delete(`v1/packwiz/pack/${slug}`)
}

export async function unArchivePack(slug: string) {
  return apiClient.patch(`v1/packwiz/pack/${slug}/unarchive`)
}

export async function publishPack(slug: string) {
  return apiClient.patch(`v1/packwiz/pack/${slug}/publish`)
}

export async function convertPackToDraft(slug: string) {
  return apiClient.patch(`v1/packwiz/pack/${slug}/draft`)
}

export async function makePackPublic(slug: string) {
  return apiClient.patch(`v1/packwiz/pack/${slug}/public`)
}

export async function makePackPrivate(slug: string) {
  return apiClient.patch(`v1/packwiz/pack/${slug}/private`)
}
