import {Type} from "class-transformer";
import type {LoaderVersions} from "@/stores/cache.ts";


export class Packs {
  @Type(() => Pack)
  packs!: Pack[];
}

export class Pack {
  slug!: string;
  name!: string;
  description!: string;
  createdAt!: string;
  createdBy!: string;
  updatedAt!: string;
  updatedBy!: string;
  deletedAt?: string;
  isPublic!: boolean;
  status!: PackStatus;
  mcVersion!: string;
  loader!: keyof LoaderVersions;
  loaderVersion!: string;
  acceptableGameVersions?: string[];

  mods?: Mod[]

  version!: string;
  packFormat!: string;

  isArchived!: boolean;
  permission!: PackPermission;
}

export enum PackStatus {
  PUBLISHED = 'published',
  DRAFT = 'draft',
}

export enum PackPermission {
  STATIC = 1,
  VIEW = 10,
  EDIT = 20,
}

export class Mod {
  id!: number;
  name!: string;
  displayName!: string;
  type!: string;
  packSlug!: string;
  fileName!: string;
  side!: "client" | "server" | "both";
  pinned!: boolean;

  source!: string;
  modKey!: string;
  versionKey!: string;

  sourceLink!: string;
}

export class PackData {
  name!: string;
  packFormat!: string;
  version!: string;
  @Type(() => Versions)
  versions!: Versions;
  options!: Options;
}

export class Versions {
  minecraft!: string;
  @Type(() => Loader)
  loader!: Loader;
}

export class Options {
  acceptableGameVersions?: string[];
}

export class Loader {
  type!: string;
  version!: string;
}

class ModSource {
  type!: string;
  modId: string;
  version: string;
  fileId: string;
  projectId: string;
}

export class ModData {
  name!: string;
  displayName!: string;
  type!: string;
  filename!: string;
  side!: "client" | "server" | "both";
  pinned!: boolean;
  @Type(() => ModSource)
  source!: ModSource;
  sourceLink!: string;
}
