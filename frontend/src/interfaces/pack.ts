import type {User} from "@/interfaces/user";
import {Type} from "class-transformer";


export class Packs {
  packs!: Pack[];
}

export class Pack {
  slug!: string;
  description!: string;

  @Type(() => Date)
  createdAt!: string;
  createdBy!: string;

  @Type(() => Date)
  updatedAt!: string;
  isPublic!: boolean;
  users!: User[];
  packData!: PackData | null;
  modData!: ModData[] | null;
}

export class PackData {
  name!: string;
  packFormat!: string;
  version!: string;
  versions!: Versions;
}

export class Versions {
  minecraft!: string;
  loader!: Loader;
}

export class Loader {
  type!: string;
  version!: string;
}

export class ModData {
  name!: string;
}
