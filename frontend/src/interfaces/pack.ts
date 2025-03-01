import {User} from "@/interfaces/user";
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
  permission!: PackPermission;
  @Type(() => PackData)
  packData!: PackData | null;
  @Type(() => ModData)
  modData!: ModData[] | null;

  get title(): string {
    return this.packData?.name || this.slug;
  }

  get archived(): boolean {
    return !!this.deletedAt;
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
}

export class Versions {
  minecraft!: string;
  @Type(() => Loader)
  loader!: Loader;
}

export class Loader {
  type!: string;
  version!: string;
}

export class ModData {
  name!: string;
}
