

export interface MinecraftDef {
  version: string;
  latest: boolean;
  snapshot: boolean;
}

export interface LoaderDef {
  name: string;
  version: string;
  latest: boolean;
}

export interface NewPackRequest {
  slug: string;
  name: string;
  version: string;
  description: string;
  minecraftVersion: string;
  loaderName: string;
  loaderVersion: string;
  acceptableVersions: string[]
}

export interface EditPackRequest {
  name?: string;
  version?: string;
  description?: string;
  minecraft?: MinecraftDef
  loader?: LoaderDef
  acceptableVersions?: string[]
}

export interface MigratePackRequest {
  minecraft: MinecraftDef;
  loader: LoaderDef;
  updateMods: boolean;
  useRecommended: boolean;
  acceptableVersions?: string[];
}

export interface AddModRequest {
  curseforge?: {
    url: string;
  }
  modrinth?: {
    url: string;
  }
  github?: {
    url: string;
  }
}

export interface ChangeModSideRequest {
  side: "client" | "server" | "both";
}

export interface ChangeModOptionRequest {
  optional: boolean;
  description: string;
  default: boolean;
}

export interface RehashRequest {
  format: "sha1" | "sha256" | "sha512";
}

export interface OidcProviderRequest {
  slug: string;
  displayName: string;
  issuerUrl: string;
  clientId: string;
  // write-only; blank on update keeps the stored secret
  clientSecret: string;
  scopes: string;
  enabled: boolean;
  autoCreateUsers: boolean;
  linkByEmail: boolean;
}

export interface OidcTestDiscoveryRequest {
  issuerUrl: string;
}

export interface CloneSnapshotRequest {
  slug: string;
  name: string;
}

export interface ImportPackRequest {
  url: string;
}
