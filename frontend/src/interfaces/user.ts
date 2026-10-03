import {Type} from "class-transformer";

export class RoleRef {
  id!: number;
  name!: string;
}

export class SystemRole {
  id!: number;
  name!: string;
  description!: string;
  scope!: string;
  assignable!: boolean;
  permissions!: string[];
}

export class User {
  id!: number;
  username!: string;
  fullName!: string;
  email!: string;
  // only sent on the signed-in user's own record (GET v1/user)
  hasPassword?: boolean;
  isSuperuser?: boolean;
  /** only sent on the signed-in user's own record (GET v1/user) */
  permissions?: string[];
  /** only sent by the admin user endpoints */
  roles?: RoleRef[];
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
