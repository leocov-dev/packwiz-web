import {apiClient} from "@/services/api.service.ts";
import type {ChangelistEntry, PackChangelist} from "@/interfaces/changelist.ts";

export async function fetchPackChangelist(packId: number): Promise<PackChangelist> {
  const response = await apiClient.get(`v1/packwiz/pack/${packId}/changelist`)
  return response.data as PackChangelist
}

/** What publishing the pack would release, against the last published state. */
export async function fetchPendingChanges(packId: number): Promise<ChangelistEntry> {
  const response = await apiClient.get(`v1/packwiz/pack/${packId}/changelist/pending`)
  return response.data as ChangelistEntry
}

export async function fetchPublicChangelist(slug: string): Promise<PackChangelist> {
  const response = await apiClient.get(`v1/public/packs/${encodeURIComponent(slug)}/changelist`)
  return response.data as PackChangelist
}
