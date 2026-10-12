import {plainToInstance} from "class-transformer";
import {CompatResponse, CompatScanResponse} from "@/interfaces/compat.ts";
import {apiClient} from "@/services/api.service";

export async function fetchCompat(packId: number): Promise<CompatResponse> {
  const response = await apiClient.get(`v1/packwiz/pack/${packId}/compat`)
  return plainToInstance(CompatResponse, response.data)
}

/** Queues a compatibility scan (202) or joins a running one (200); poll {@link fetchCompat}. */
export async function requestCompatScan(packId: number): Promise<CompatScanResponse> {
  const response = await apiClient.post(`v1/packwiz/pack/${packId}/compat/scan`)
  return plainToInstance(CompatScanResponse, response.data)
}
