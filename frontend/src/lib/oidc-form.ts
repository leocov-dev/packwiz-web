// Pure validation/normalisation helpers for the OIDC provider form. They
// mirror the backend rules in dto/oidc_provider_form.go so the form can give
// feedback before a round trip; the backend remains the source of truth.

export const DEFAULT_OIDC_SCOPES = 'openid profile email'

const SLUG_PATTERN = /^[a-z0-9]([a-z0-9-]{0,62}[a-z0-9])?$/
const LOOPBACK_HOSTS = ['localhost', '127.0.0.1', '[::1]', '::1']

export function slugify(name: string): string {
  return name
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-+|-+$/g, '')
    .slice(0, 64)
    .replace(/-+$/g, '')
}

export function validateSlug(slug: string): true | string {
  if (!slug) return 'Slug is required.'
  if (!SLUG_PATTERN.test(slug)) {
    return 'Use 1-64 lowercase letters, digits or hyphens; no leading or trailing hyphen.'
  }
  return true
}

export function validateIssuerUrl(raw: string): true | string {
  const value = raw.trim()
  if (!value) return 'Issuer URL is required.'

  let url: URL
  try {
    url = new URL(value)
  } catch {
    return 'Must be an absolute URL.'
  }

  if (url.search || url.hash) return 'Must not contain a query or fragment.'
  if (url.protocol === 'https:') return true
  if (url.protocol === 'http:' && LOOPBACK_HOSTS.includes(url.hostname)) return true
  return 'Must use https.'
}

export function normalizeScopes(raw: string): string {
  const scopes = [...new Set(raw.split(/\s+/).filter(Boolean))]
  return scopes.length ? scopes.join(' ') : DEFAULT_OIDC_SCOPES
}

export function validateScopes(raw: string): true | string {
  return normalizeScopes(raw).split(' ').includes('openid') || "Scopes must include 'openid'."
}

export type ProviderStatus = 'enabled' | 'disabled' | 'broken'

export function providerStatus(p: { enabled: boolean, broken: boolean }): ProviderStatus {
  if (p.broken) return 'broken'
  return p.enabled ? 'enabled' : 'disabled'
}

export function deleteConfirmationText(displayName: string, orphanedUsers: number): string {
  let text = `Delete "${displayName}"? Everyone's linked accounts for this provider will be removed.`
  if (orphanedUsers > 0) {
    const who = orphanedUsers === 1 ? '1 user has' : `${orphanedUsers} users have`
    text += `\n\n${who} no password and no other linked account. ` +
      `They will be locked out until an admin sets them up again.`
  }
  return text
}
