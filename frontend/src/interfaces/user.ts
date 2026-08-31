import {Type} from "class-transformer";

export class User {
  id!: number;
  username!: string;
  fullName!: string;
  email!: string;
  identityProvider!: string;
  isAdmin!: boolean;
  isActive!: boolean;
  createdAt!: string;
  updatedAt!: string;
}

export class Pagination {
  page!: number;
  size!: number;
  total!: number;
}

export class UserListResponse {
  @Type(() => User)
  results!: User[];
  @Type(() => Pagination)
  pagination!: Pagination;
}
