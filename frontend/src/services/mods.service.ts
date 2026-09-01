import {apiClient} from "@/services/api.service.ts";
import type {AddModRequest, ChangeModOptionRequest, ChangeModSideRequest} from "@/interfaces/requests.ts";
import {Mod, ModDependenciesResponse, ModSearchResponse} from "@/interfaces/pack.ts";
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

export async function changeModSide(packId: number, modId: number, request: ChangeModSideRequest) {
  return apiClient.patch(`v1/packwiz/pack/${packId}/mod/${modId}/side`, request)
}

export async function changeModOption(packId: number, modId: number, request: ChangeModOptionRequest) {
  return apiClient.patch(`v1/packwiz/pack/${packId}/mod/${modId}/option`, request)
}

export async function pinMod(packId: number, modId: number) {
  return apiClient.patch(`v1/packwiz/pack/${packId}/mod/${modId}/pin`)
}

export async function unpinMod(packId: number, modId: number) {
  return apiClient.patch(`v1/packwiz/pack/${packId}/mod/${modId}/unpin`)
}

export async function updateModFromSource(packId: number, modId: number) {
  return apiClient.patch(`v1/packwiz/pack/${packId}/mod/${modId}/update`)
}

export async function removeMod(packId: number, modId: number) {
  return apiClient.delete(`v1/packwiz/pack/${packId}/mod/${modId}`)
}

export async function searchModrinthMods(packId: number, query: string, versions?: string[]): Promise<ModSearchResponse> {
  const params = new URLSearchParams({q: query})
  for (const v of versions || []) {
    params.append('versions', v)
  }
  const response = await apiClient.get(`v1/packwiz/pack/${packId}/mod/search?${params.toString()}`)
  return plainToInstance(ModSearchResponse, response.data)
}
