// Pure helpers for the pack webhook form. They mirror the backend rules in
// dto/pack_webhook.go so the form can give feedback before a round trip; the
// backend remains the source of truth.
import type {PackWebhookRequest, PackWebhookResponse, WebhookMethod} from "@/interfaces/webhook.ts"

export const WEBHOOK_METHODS: WebhookMethod[] = ['POST', 'PUT', 'PATCH', 'GET']

const HEADER_NAME = /^[!#$%&'*+.^_`|~0-9A-Za-z-]{1,128}$/
const FORBIDDEN_HEADERS = ['host', 'content-length', 'transfer-encoding', 'connection']
const MAX_HEADERS = 20

export interface WebhookHeaderRow {
  name: string
  value: string
  /** a value is stored on the server; leaving the value blank keeps it */
  stored: boolean
  /** show the value instead of masking it; UI only, never sent */
  visible: boolean
}

export interface WebhookFormData {
  enabled: boolean
  url: string
  method: WebhookMethod
  headers: WebhookHeaderRow[]
}

export function formFromResponse(res: PackWebhookResponse): WebhookFormData {
  return {
    enabled: res.enabled,
    url: res.url,
    method: res.method || 'POST',
    headers: res.headers.map(h => ({name: h.name, value: '', stored: true, visible: false})),
  }
}

export function validateWebhookUrl(raw: string): true | string {
  const value = raw.trim()
  if (!value) return 'URL is required.'

  let url: URL
  try {
    url = new URL(value)
  } catch {
    return 'Must be an absolute URL.'
  }
  if (url.protocol !== 'http:' && url.protocol !== 'https:') return 'Must use http or https.'
  if (url.username || url.password) return 'Do not put credentials in the URL; use a header.'
  return true
}

/** Returns the first problem with the form, or null when it can be saved. */
export function validateWebhookForm(form: WebhookFormData): string | null {
  if (form.url.trim() || form.enabled) {
    const url = validateWebhookUrl(form.url)
    if (url !== true) return url
  }

  const rows = form.headers.filter(h => h.name.trim() || h.value)
  if (rows.length > MAX_HEADERS) return `At most ${MAX_HEADERS} headers are allowed.`

  const seen = new Set<string>()
  for (const row of rows) {
    const name = row.name.trim()
    if (!HEADER_NAME.test(name)) return `Invalid header name "${name}".`
    const key = name.toLowerCase()
    if (FORBIDDEN_HEADERS.includes(key)) return `Header "${name}" cannot be set.`
    if (seen.has(key)) return `Duplicate header "${name}".`
    seen.add(key)
    if (!row.value && !row.stored) return `Header "${name}" needs a value.`
    if (/[\r\n]/.test(row.value)) return `Invalid value for header "${name}".`
  }
  return null
}

export function buildWebhookRequest(form: WebhookFormData): PackWebhookRequest {
  return {
    enabled: form.enabled,
    url: form.url.trim(),
    method: form.method,
    headers: form.headers
      .filter(h => h.name.trim() || h.value)
      .map(h => ({name: h.name.trim(), value: h.value})),
  }
}

/** One line describing the last delivery attempt. */
export function lastAttemptSummary(res: Pick<PackWebhookResponse, 'lastAttemptAt' | 'lastError' | 'lastStatusCode'>): string {
  if (!res.lastAttemptAt) return 'Not called yet.'
  const when = new Date(res.lastAttemptAt).toLocaleString()
  if (res.lastError) return `Failed ${when}: ${res.lastError}`
  return `Succeeded ${when} (HTTP ${res.lastStatusCode}).`
}
