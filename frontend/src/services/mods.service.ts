import {apiClient} from "@/services/api.service.ts";
import type {AddModRequest} from "@/interfaces/requests.ts";
import {Mod, ModDependenciesResponse} from "@/interfaces/pack.ts";
import {plainToInstance} from "class-transformer";

export async function fetchOneMod(slug: string, modSlug: string): Promise<Mod> {
  const response = await apiClient.get(`v1/packwiz/pack/${slug}/mod/${modSlug}`)
  return plainToInstance(Mod, response.data)

}

export async function addMod(slug: string, addModRequest: AddModRequest) {
  return apiClient.post(`v1/packwiz/pack/${slug}/mod`, addModRequest)
}

export async function listMissingDependencies(slug: string, addModRequest: AddModRequest): Promise<ModDependencies> {
  const response = await apiClient.post(`v1/packwiz/pack/${slug}/mod/missing-dependencies`, addModRequest)
  return plainToInstance(ModDependenciesResponse, response.data)
}
