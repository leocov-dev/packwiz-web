import {Type} from "class-transformer";
import type {LoaderVersions} from "@/stores/cache.ts";
import type {User} from "@/interfaces/user.ts";


export class AllPacksResponse {
  @Type(() => PackResponse)
  packs?: PackResponse[];
}

export class Pack {
  id!: number;
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
  /** Effective permission names on this pack. */
  permissions: string[] = [];
  /** Pack role name, display only; empty when access comes from a system role. */
  currentUserRole?: string;
}

export enum PackStatus {
  PUBLISHED = 'published',
  DRAFT = 'draft',
}

export class Mod {
  id!: number;
  slug!: string;
  packId!: string;
  name!: string;
  type!: string;
  fileName!: string;
  /** Installed version; DB-only, absent for legacy mods until their next update. */
  version?: string;
  side!: "client" | "server" | "both";
  pinned!: boolean;
  source!: string;
  update!: {[key: string]: string | number}
  createdBy!: number;
  createdAt!: string;
  updatedBy!: number;
  updatedAt!: string;
  isDependency!: boolean;
  dependencyIds?: number[];
  option!: {optional: boolean; description: string; default: boolean};
}


export class MigrateDryRunMod {
  modId!: number;
  slug!: string;
  name!: string;
  pinned!: boolean;
  updateAvailable!: boolean;
  updateString?: string;
  incompatible!: boolean;
  error?: string;
}

export class MigrateDryRunResponse {
  @Type(() => MigrateDryRunMod)
  mods!: MigrateDryRunMod[];
}


export class MigrateResponse {
  modsQueued!: boolean;
  jobId?: number;
}

export class MigrateJobStatusResponse {
  state!: "available" | "running" | "completed" | "discarded" | "cancelled" | "retryable" | "pending" | "scheduled";
  @Type(() => MigrateDryRunMod)
  mods?: MigrateDryRunMod[];
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


export class ModSearchResult {
  slug!: string;
  title!: string;
  description!: string;
  iconUrl!: string;
  projectId!: string;
  installed?: boolean;
  author?: string;
  downloads?: number;
  loaders?: string[];
  categories?: string[];
}

export class ModSearchResponse {
  @Type(() => ModSearchResult)
  results?: ModSearchResult[];
}


export class PackCollaborator {
  userId!: number;
  username!: string;
  fullName!: string;
  email!: string;
  roleId!: number;
  roleName!: string;
  isActive!: boolean;
  createdAt!: string;
}

export class PackRole {
  id!: number;
  name!: string;
  description!: string;
  permissions!: string[];
}

export class PackRolesResponse {
  @Type(() => PackRole)
  roles?: PackRole[];
}

export class PackCollaboratorsResponse {
  @Type(() => PackCollaborator)
  users?: PackCollaborator[];
}

export class UserSearchResult {
  userId!: number;
  username!: string;
  fullName!: string;
  email!: string;
}

export class UserSearchResponse {
  @Type(() => UserSearchResult)
  users?: UserSearchResult[];
}

export class UpdateModResponse {
  updated!: boolean;
}

export class UpdateAllItem {
  modId!: number;
  slug!: string;
  name!: string;
  fileName?: string;
  reason?: string;
  error?: string;
}

export class UpdateAllResponse {
  @Type(() => UpdateAllItem)
  updated!: UpdateAllItem[];
  @Type(() => UpdateAllItem)
  skipped!: UpdateAllItem[];
  @Type(() => UpdateAllItem)
  failed!: UpdateAllItem[];
  upToDate!: number;
  notChecked!: number;
}

export type UpdateCheckStatus = "idle" | "queued" | "running" | "done" | "failed"

export class UpdateCheckJobResponse {
  jobId!: number;
  status!: UpdateCheckStatus;
}

export class UpdateCheckItem {
  modId!: number;
  updateAvailable!: boolean;
  updateString?: string;
  /** Latest available version, when the source reports one. */
  latestVersion?: string;
  error?: string;
}

export class UpdateChecksResponse {
  status!: UpdateCheckStatus;
  /** When the returned results were produced; kept after a failed run. */
  checkedAt?: string | null;
  /** When the latest run finished, successfully or not. */
  runFinishedAt?: string | null;
  error?: string;
  @Type(() => UpdateCheckItem)
  results!: UpdateCheckItem[];
  availableCount!: number;
}

/** Outcome of importing a live packwiz pack from a pack.toml url. */
export class ImportPackResponse {
  packId!: number;
  slug!: string;
  name!: string;
  modsImported!: number;
  skippedMods: string[] = [];
  skippedFiles: string[] = [];
  warnings: string[] = [];
}


export class PublicPackMod {
  slug!: string;
  name!: string;
  type!: string;
  side!: "client" | "server" | "both";
  version?: string;
  source!: string;
  optional!: boolean;
  isDependency!: boolean;
}

/** Unauthenticated view of a public, published pack. */
export class PublicPack {
  slug!: string;
  name!: string;
  description!: string;
  author!: string;
  version!: string;
  mcVersion!: string;
  loader!: string;
  loaderVersion!: string;
  acceptableGameVersions!: string[];
  packFormat!: string;
  updatedAt!: string;
  packTomlUrl!: string;
  @Type(() => PublicPackMod)
  mods!: PublicPackMod[];
}
