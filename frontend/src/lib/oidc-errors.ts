// Maps the generic error codes the backend puts in `?error=` (login page) and
// `?oidcError=` (profile page) onto messages. The backend never forwards IdP
// error text, so unknown codes fall back to a generic message.

const MESSAGES: Record<string, string> = {
  oidc_failed: 'Sign-in with the provider failed. Please try again.',
  oidc_not_permitted: 'This account is not permitted to sign in with that provider. Ask an administrator for access.',
  oidc_unavailable: 'That sign-in provider is not available right now.',
  oidc_already_linked: 'That account is already linked to a user.',
}

const GENERIC = 'Sign-in failed. Please try again.'

function firstValue(raw: unknown): string | null {
  const value = Array.isArray(raw) ? raw[0] : raw
  return typeof value === 'string' && value ? value : null
}

export function oidcErrorMessage(raw: unknown): string | null {
  const code = firstValue(raw)
  if (!code) return null
  return Object.prototype.hasOwnProperty.call(MESSAGES, code) ? MESSAGES[code] : GENERIC
}
