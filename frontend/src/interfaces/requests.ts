

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
  minecraft: MinecraftDef
  loader: LoaderDef
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
