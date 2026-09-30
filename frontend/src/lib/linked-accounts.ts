import type {OidcPublicProvider, UserIdentity} from "@/interfaces/oidc.ts"

// Providers the user can still link: enabled ones with no identity yet. The
// backend allows one identity per user per provider.
export function linkableProviders(
  providers: OidcPublicProvider[],
  identities: UserIdentity[],
): OidcPublicProvider[] {
  const linked = new Set(identities.map(i => i.providerSlug))
  return providers.filter(p => !linked.has(p.slug))
}
