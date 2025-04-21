import {Type} from "class-transformer";


export class Packs {
  @Type(() => Pack)
  packs!: Pack[];
}

export class Pack {
  slug!: string;
  description!: string;
  createdAt!: string;
  createdBy!: string;
  updatedAt!: string;
  deletedAt?: string;
  isPublic!: boolean;
  status!: PackStatus;
  isArchived!: boolean;
  permission!: PackPermission;
  @Type(() => PackData)
  packData!: PackData | null;
  @Type(() => ModData)
  modData!: ModData[] | null;

  get title(): string {
    return this.packData?.name || this.slug;
  }

  get dataMissing(): boolean {
    return !this.packData
  }
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
