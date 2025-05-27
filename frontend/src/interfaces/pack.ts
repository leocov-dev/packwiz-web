import {Type} from "class-transformer";
import type {LoaderVersions} from "@/stores/cache.ts";
import type {User} from "@/interfaces/user.ts";


export class AllPacksResponse {
  @Type(() => PackResponse)
  packs?: PackResponse[];
}

export class Pack {
  slug!: string;
  name!: string;
  description!: string;
  createdAt!: string;
  createdBy!: number;
  author!: User;
  updatedAt!: string;
  updatedBy!: number;
  deletedAt?: string;
  isPublic!: boolean;
  status!: PackStatus;
  mcVersion!: string;
  loader!: keyof LoaderVersions;
  loaderVersion!: string;
  acceptableGameVersions?: string[];
  version!: string;
  packFormat!: string;
  mods?: Mod[]

  get isArchived(): boolean {
    return this.deletedAt !== null && this.deletedAt !== undefined;
  }
}

export class PackResponse extends Pack {
  currentUserPermission!: PackPermission;
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
  packSlug!: string;
  slug!: string;
  name!: string;
  type!: string;
  fileName!: string;
  side!: "client" | "server" | "both";
  pinned!: boolean;
  source!: string;
  update!: {[key: string]: string | number}
  createdBy!: number;
  createdAt!: string;
  updatedBy!: number;
  updatedAt!: string;
}


export class ModDependency {
  fileName!: string;
  modType!: string;
  name!: string;
  side!: "client" | "server" | "both";
  slug!: string;
  url!: string;
}


export class ModDependenciesResponse {
  missing!: ModDependency[]
}
