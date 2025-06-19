import {apiClient} from "@/services/api.service.ts";
import type {AddModRequest} from "@/interfaces/requests.ts";
import {Mod, ModDependenciesResponse} from "@/interfaces/pack.ts";
import {plainToInstance} from "class-transformer";

export async function fetchOneMod(packId: number, modId: number): Promise<Mod> {
  const response = await apiClient.get(`v1/packwiz/pack/${packId}/mod/${modId}`)
  return plainToInstance(Mod, response.data)

}

export async function addMod(packId: number, addModRequest: AddModRequest) {
  return apiClient.post(`v1/packwiz/pack/${packId}/mod`, addModRequest)
}

export async function listMissingDependencies(packId: number, addModRequest: AddModRequest): Promise<ModDependenciesResponse> {
  const response = await apiClient.post(`v1/packwiz/pack/${packId}/mod/missing-dependencies`, addModRequest)
  return plainToInstance(ModDependenciesResponse, response.data)
}
