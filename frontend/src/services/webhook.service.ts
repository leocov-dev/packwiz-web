import {apiClient} from "@/services/api.service.ts"
import type {
  PackWebhookRequest,
  PackWebhookResponse,
  PackWebhookTestResponse,
} from "@/interfaces/webhook.ts"

export async function getPackWebhook(packId: number): Promise<PackWebhookResponse> {
  const response = await apiClient.get(`v1/packwiz/pack/${packId}/webhook`)
  return response.data
}

export async function savePackWebhook(packId: number, request: PackWebhookRequest): Promise<PackWebhookResponse> {
  const response = await apiClient.put(`v1/packwiz/pack/${packId}/webhook`, request)
  return response.data
}

export async function testPackWebhook(packId: number): Promise<PackWebhookTestResponse> {
  const response = await apiClient.post(`v1/packwiz/pack/${packId}/webhook/test`)
  return response.data
}
