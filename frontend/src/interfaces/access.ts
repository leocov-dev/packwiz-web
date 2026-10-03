import {Type} from "class-transformer";
import {Pagination} from "@/interfaces/audit.ts";

export class AccessDay {
  date!: string;
  success!: number;
  failure!: number;
}

export class AccessUserCount {
  userId!: number;
  username!: string;
  count!: number;
}

export class AccessIpCount {
  ipAddress!: string;
  count!: number;
}

export class AccessPackCount {
  packId!: number;
  name!: string;
  slug!: string;
  success!: number;
  failure!: number;
}

export class AccessRecord {
  id!: number;
  createdAt!: string;
  packId!: number | null;
  packSlug!: string;
  packName!: string;
  userId!: number | null;
  username!: string;
  ipAddress!: string;
  userAgent!: string;
  statusCode!: number;
  success!: boolean;
}

export class PackAccessSummary {
  days!: number;
  @Type(() => AccessDay)
  series!: AccessDay[];
  total!: number;
  uniqueIps!: number;
  uniqueUsers!: number;
  @Type(() => AccessUserCount)
  topUsers!: AccessUserCount[];
  @Type(() => AccessIpCount)
  topIps!: AccessIpCount[];
}

export class SystemAccessSummary {
  days!: number;
  @Type(() => AccessDay)
  series!: AccessDay[];
  success!: number;
  failure!: number;
  uniqueIps!: number;
  @Type(() => AccessPackCount)
  packs!: AccessPackCount[];
  @Type(() => AccessIpCount)
  topFailedIps!: AccessIpCount[];
}

export class AccessRecordListResponse {
  @Type(() => AccessRecord)
  results!: AccessRecord[];
  @Type(() => Pagination)
  pagination!: Pagination;
}

export type AccessOutcome = "all" | "success" | "failure"
