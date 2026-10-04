import axios from "axios"
import {
  AllPacksResponse,
  MigrateDryRunResponse,
  MigrateJobStatusResponse,
  MigrateResponse,
  ImportPackResponse,
  Pack,
  PackCollaboratorsResponse,
  PackRolesResponse,
  PackResponse,
  PublicPack,
  UpdateAllResponse,
  UpdateCheckJobResponse,
  UpdateChecksResponse,
  UserSearchResponse
} from "@/interfaces/pack";
import {apiClient} from "@/services/api.service";
import {plainToInstance} from "class-transformer";
import {writeToClipboard} from "@/lib/clipboard";
import type {EditPackRequest, ImportPackRequest, MigratePackRequest, NewPackRequest, RehashRequest} from "@/interfaces/requests.ts";


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


export interface PackLinks {
  /** pack.toml link for packwiz installers. */
  link: string
  /** Importable MultiMC / Prism instance zip; also works as an import URL. */
  multimcLink: string
}

export async function getPackLinks(packId: number): Promise<PackLinks> {
  const response = await apiClient.get(`v1/packwiz/pack/${packId}/link`);
  return response.data as PackLinks
}

export async function getPackPublicLink(packId: number): Promise<string> {
  return (await getPackLinks(packId)).link
}

/** Copies the instance zip URL, for a launcher's "import from URL". */
export async function instanceUrlToClipboard(packId: number) {
  await writeToClipboard((await getPackLinks(packId)).multimcLink)
}

/** Downloads the MultiMC / Prism instance for a pack the user can access. */
export async function downloadMultiMCInstance(packId: number) {
  await downloadInstanceZip((await getPackLinks(packId)).multimcLink)
}


export async function linkToClipboard(packId: number) {
  const link = await getPackPublicLink(packId)
  await writeToClipboard(link)
}

export async function openPublicLink(packId: number) {
  const link = await getPackPublicLink(packId)
  window.open(link, '_blank')
}

// See https://packwiz.infra.link/tutorials/installing/packwiz-installer/
export function getClientSetupCommand(link: string): string {
  return `"$INST_JAVA" -jar packwiz-installer-bootstrap.jar ${link}`
}

export async function clientSetupCommandToClipboard(packId: number) {
  const link = await getPackPublicLink(packId)
  await writeToClipboard(getClientSetupCommand(link))
}

/**
 * Downloads a MultiMC / Prism Launcher instance zip from its consumer URL.
 * The URL carries its own token, so plain axios is used: no session, and no
 * api interceptor logging the user out on a 401. The server rate-limits this,
 * so a 429 gets its own message.
 */
export async function downloadInstanceZip(url: string) {
  let response
  try {
    response = await axios.get(url, {responseType: 'blob'})
  } catch (e) {
    if (axios.isAxiosError(e) && e.response?.status === 429) {
      throw new Error('Too many downloads, try again in a minute')
    }
    throw new Error('Could not generate the instance')
  }

  const objectUrl = URL.createObjectURL(response.data)
  const a = document.createElement('a')
  a.href = objectUrl
  // the URL ends in "<Pack Name>.zip"; launchers name the instance after it
  a.download = decodeURIComponent(new URL(url).pathname.split('/').pop() ?? 'instance.zip')
  document.body.appendChild(a)
  a.click()
  a.remove()
  // revoking right after click can cancel the download in Safari
  setTimeout(() => URL.revokeObjectURL(objectUrl), 60_000)
}

/** Path of the standalone public page for a public pack. */
export function publicPackPath(slug: string): string {
  return `/p/${encodeURIComponent(slug)}`
}

export async function fetchPublicPack(slug: string): Promise<PublicPack> {
  const response = await apiClient.get(`v1/public/packs/${encodeURIComponent(slug)}`)
  return plainToInstance(PublicPack, response.data)
}

export async function newPack(request: NewPackRequest): Promise<Pack> {
  const response = await apiClient.post('v1/packwiz/pack', request)
  return plainToInstance(Pack, response.data)
}

export async function importPack(request: ImportPackRequest): Promise<ImportPackResponse> {
  const response = await apiClient.post('v1/packwiz/import', request)
  return plainToInstance(ImportPackResponse, response.data)
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

// update-all checks every mod against remote APIs; allow far longer than the global default
const UPDATE_ALL_TIMEOUT_MS = 5 * 60 * 1000

export async function updateAllMods(packId: number): Promise<UpdateAllResponse> {
  const response = await apiClient.patch(
    `v1/packwiz/pack/${packId}/update-all`,
    undefined,
    {timeout: UPDATE_ALL_TIMEOUT_MS},
  )
  return plainToInstance(UpdateAllResponse, response.data)
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

export async function listPackRoles(): Promise<PackRolesResponse> {
  const response = await apiClient.get('v1/packwiz/pack/roles');
  return plainToInstance(PackRolesResponse, response.data)
}

export async function addPackCollaborator(packId: number, userId: number, roleId: number) {
  return apiClient.post(`v1/packwiz/pack/${packId}/users`, {userId, roleId})
}

export async function updateCollaboratorRole(packId: number, userId: number, roleId: number) {
  return apiClient.patch(`v1/packwiz/pack/${packId}/users/${userId}`, {roleId})
}

export async function removeCollaborator(packId: number, userId: number) {
  return apiClient.delete(`v1/packwiz/pack/${packId}/users/${userId}`)
}

/** Starts (or joins) an async update check job; poll {@link fetchUpdateChecks} for results. */
export async function checkForUpdates(packId: number, force: boolean = false): Promise<UpdateCheckJobResponse> {
  const response = await apiClient.post(
    `v1/packwiz/pack/${packId}/updates/check`,
    undefined,
    {params: force ? {force: 'true'} : undefined},
  )
  return plainToInstance(UpdateCheckJobResponse, response.data)
}

export async function fetchUpdateChecks(packId: number): Promise<UpdateChecksResponse> {
  const response = await apiClient.get(`v1/packwiz/pack/${packId}/updates`)
  return plainToInstance(UpdateChecksResponse, response.data)
}
