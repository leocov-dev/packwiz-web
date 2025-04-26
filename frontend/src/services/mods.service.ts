import {apiClient} from "@/services/api.service.ts";
import type {AddModRequest} from "@/interfaces/requests.ts";
import {ModData} from "@/interfaces/pack.ts";
import {plainToInstance} from "class-transformer";

export async function fetchOneMod(slug: string, modSlug: string): Promise<ModData> {
  const response = await apiClient.get(`v1/packwiz/pack/${slug}/mod/${modSlug}`)
  return plainToInstance(ModData, response.data)

}

export async function addMod(slug: string, addModRequest: AddModRequest) {
  return apiClient.post(`v1/packwiz/pack/${slug}/mod`, addModRequest)
}
