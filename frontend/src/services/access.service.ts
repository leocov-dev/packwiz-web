import {plainToInstance} from "class-transformer";
import {apiClient} from "@/services/api.service.ts";
import {
  AccessDay,
  type AccessOutcome,
  AccessRecordListResponse,
  PackAccessSummary,
  SystemAccessSummary,
} from "@/interfaces/access.ts";

function packAccessUrl(packId: number, action?: string): string {
  return `v1/packwiz/pack/${packId}/access${action ? `/${action}` : ''}`
}

export async function fetchPackAccessSeries(packId: number, days: number): Promise<AccessDay[]> {
  const response = await apiClient.get(`${packAccessUrl(packId, 'series')}?days=${days}`)
  return plainToInstance(AccessDay, response.data as unknown[])
}

export async function fetchPackAccessSummary(packId: number, days: number): Promise<PackAccessSummary> {
  const response = await apiClient.get(`${packAccessUrl(packId)}?days=${days}`)
  return plainToInstance(PackAccessSummary, response.data)
}

export async function fetchPackAccessRecent(
  packId: number,
  days: number,
  page: number,
  pageSize: number,
): Promise<AccessRecordListResponse> {
  const params = new URLSearchParams({days: String(days), page: String(page), pageSize: String(pageSize)})
  const response = await apiClient.get(`${packAccessUrl(packId, 'recent')}?${params.toString()}`)
  return plainToInstance(AccessRecordListResponse, response.data)
}

/**
 * System-wide access sources: pack.toml syncs, or MultiMC / Prism instance zip
 * downloads. Download stats are only shown to system admins.
 */
export type AccessSource = 'pack-access' | 'instance-downloads'

export async function fetchSystemAccessSummary(
  days: number,
  source: AccessSource = 'pack-access',
): Promise<SystemAccessSummary> {
  const response = await apiClient.get(`v1/admin/${source}?days=${days}`)
  return plainToInstance(SystemAccessSummary, response.data)
}

export async function fetchSystemAccessRecent(
  days: number,
  page: number,
  pageSize: number,
  outcome: AccessOutcome,
  packId?: number,
  source: AccessSource = 'pack-access',
): Promise<AccessRecordListResponse> {
  const params = new URLSearchParams({
    days: String(days),
    page: String(page),
    pageSize: String(pageSize),
    outcome,
  })
  if (packId) {
    params.set('packId', String(packId))
  }
  const response = await apiClient.get(`v1/admin/${source}/recent?${params.toString()}`)
  return plainToInstance(AccessRecordListResponse, response.data)
}
