export class OidcProvider {
  id!: number;
  slug!: string;
  displayName!: string;
  issuerUrl!: string;
  clientId!: string;
  hasSecret!: boolean;
  scopes!: string;
  enabled!: boolean;
  autoCreateUsers!: boolean;
  linkByEmail!: boolean;
  redirectUri!: string;
  broken!: boolean;
  brokenError?: string;
}

export class OidcPublicProvider {
  slug!: string;
  displayName!: string;
}

export class OidcDiscoveryResult {
  issuer!: string;
  authorizationEndpoint!: string;
  tokenEndpoint!: string;
}

export class OrphanedUsers {
  count!: number;
}

export class UserIdentity {
  id!: number;
  providerId!: number;
  providerSlug!: string;
  providerName!: string;
  email!: string;
  createdAt!: string;
  lastLoginAt!: string | null;
}

export class LinkIdentityResult {
  redirectUrl!: string;
}
