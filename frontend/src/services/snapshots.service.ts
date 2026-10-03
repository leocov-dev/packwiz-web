import {plainToInstance} from "class-transformer";
import {apiClient} from "@/services/api.service.ts";
import {PackResponse} from "@/interfaces/pack.ts";
import {
  PackSnapshotDetailResponse,
  PackSnapshotListResponse,
  RevertSnapshotResponse,
  type SnapshotAgainst,
} from "@/interfaces/snapshot.ts";
import type {CloneSnapshotRequest} from "@/interfaces/requests.ts";

function snapshotsUrl(packId: number, snapshotId?: number, action?: string): string {
  let url = `v1/packwiz/pack/${packId}/snapshots`
  if (snapshotId !== undefined) {
    url += `/${snapshotId}`
  }
  if (action) {
    url += `/${action}`
  }
  return url
}

export async function fetchPackSnapshots(
  packId: number,
  page: number,
  pageSize: number,
  includeAbandoned: boolean,
): Promise<PackSnapshotListResponse> {
  const params = new URLSearchParams({
    page: String(page),
    pageSize: String(pageSize),
    abandoned: String(includeAbandoned),
  })
  const response = await apiClient.get(`${snapshotsUrl(packId)}?${params.toString()}`)
  return plainToInstance(PackSnapshotListResponse, response.data)
}

export async function fetchPackSnapshot(
  packId: number,
  snapshotId: number,
  against: SnapshotAgainst,
): Promise<PackSnapshotDetailResponse> {
  const response = await apiClient.get(`${snapshotsUrl(packId, snapshotId)}?against=${against}`)
  return plainToInstance(PackSnapshotDetailResponse, response.data)
}

export async function revertToSnapshot(packId: number, snapshotId: number): Promise<RevertSnapshotResponse> {
  const response = await apiClient.post(snapshotsUrl(packId, snapshotId, "revert"))
  return plainToInstance(RevertSnapshotResponse, response.data)
}

export async function cloneFromSnapshot(
  packId: number,
  snapshotId: number,
  request: CloneSnapshotRequest,
): Promise<PackResponse> {
  const response = await apiClient.post(snapshotsUrl(packId, snapshotId, "clone"), request)
  return plainToInstance(PackResponse, response.data)
}
