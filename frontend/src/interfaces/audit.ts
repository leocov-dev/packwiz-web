import {Type} from "class-transformer";

export class Audit {
  id!: number;
  createdAt!: string;
  userId!: number;
  action!: string;
  actionParams!: string;
  ipAddress!: string;
}

export class Pagination {
  page!: number;
  size!: number;
  total!: number;
}

export class AuditListResponse {
  @Type(() => Audit)
  results!: Audit[];
  @Type(() => Pagination)
  pagination!: Pagination;
}
