export type WebhookMethod = 'POST' | 'PUT' | 'PATCH' | 'GET'

export interface WebhookHeaderRequest {
  name: string
  /** blank keeps the stored value of the header with the same name */
  value: string
}

export interface PackWebhookRequest {
  enabled: boolean
  url: string
  method: WebhookMethod
  headers: WebhookHeaderRequest[]
}

export interface PackWebhookResponse {
  enabled: boolean
  url: string
  method: WebhookMethod
  /** names only, values are write-only */
  headers: { name: string }[]
  broken: boolean
  lastStatusCode: number
  lastError: string
  lastAttemptAt: string | null
  lastSuccessAt: string | null
}

export interface PackWebhookTestResponse {
  ok: boolean
  statusCode: number
  error: string
}
