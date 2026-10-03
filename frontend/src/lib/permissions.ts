// Permission names. They mirror the backend `permissions` table.
export const Perm = {
  PackCreate: "pack.create",
  UserView: "user.view",
  UserLookup: "user.lookup",
  UserCreate: "user.create",
  UserManage: "user.manage",
  UserRolesAssign: "user.roles.assign",
  AuditView: "audit.view",
  OidcManage: "oidc.manage",

  PackConsume: "pack.consume",
  PackView: "pack.view",
  PackLink: "pack.link",
  PackModAdd: "pack.mod.add",
  PackModRemove: "pack.mod.remove",
  PackModUpdate: "pack.mod.update",
  PackModConfigure: "pack.mod.configure",
  PackUpdatesCheck: "pack.updates.check",
  PackMigrate: "pack.migrate",
  PackRehash: "pack.rehash",
  PackInfoEdit: "pack.info.edit",
  PackPublish: "pack.publish",
  PackVisibility: "pack.visibility",
  PackArchive: "pack.archive",
  PackUsersView: "pack.users.view",
  PackUsersManage: "pack.users.manage",
  PackSnapshotView: "pack.snapshot.view",
  PackSnapshotRevert: "pack.snapshot.revert",
} as const

export type PermissionName = typeof Perm[keyof typeof Perm]

export function hasPermission(permissions: readonly string[] | undefined | null, name: PermissionName): boolean {
  return !!permissions && permissions.includes(name)
}
