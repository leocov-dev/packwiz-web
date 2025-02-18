import {Type} from "class-transformer";

export class User {
  id!: number;
  username!: string;
  email!: string;
  identityProvider!: string;
  isAdmin!: boolean;
  @Type(() => Date)
  createdAt!: string;
  @Type(() => Date)
  updatedAt!: string;
}
