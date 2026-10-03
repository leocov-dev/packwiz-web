package authz_svc

// Permission names. They mirror the seeded rows in the permissions table; the
// superuser is granted this in-code list so it never depends on the database.
const (
	PackCreate      = "pack.create"
	UserView        = "user.view"
	UserLookup      = "user.lookup"
	UserCreate      = "user.create"
	UserManage      = "user.manage"
	UserRolesAssign = "user.roles.assign"
	AuditView       = "audit.view"
	OidcManage      = "oidc.manage"

	PackConsume      = "pack.consume"
	PackView         = "pack.view"
	PackLink         = "pack.link"
	PackModAdd       = "pack.mod.add"
	PackModRemove    = "pack.mod.remove"
	PackModUpdate    = "pack.mod.update"
	PackModConfigure = "pack.mod.configure"
	PackUpdatesCheck = "pack.updates.check"
	PackMigrate      = "pack.migrate"
	PackRehash       = "pack.rehash"
	PackInfoEdit     = "pack.info.edit"
	PackPublish      = "pack.publish"
	PackVisibility   = "pack.visibility"
	PackArchive      = "pack.archive"
	PackUsersView    = "pack.users.view"
	PackUsersManage  = "pack.users.manage"

	PackSnapshotView   = "pack.snapshot.view"
	PackSnapshotRevert = "pack.snapshot.revert"
)

// AllPermissions lists every permission name.
var AllPermissions = []string{
	PackCreate, UserView, UserLookup, UserCreate, UserManage, UserRolesAssign, AuditView, OidcManage,
	PackConsume, PackView, PackLink, PackModAdd, PackModRemove, PackModUpdate, PackModConfigure,
	PackUpdatesCheck, PackMigrate, PackRehash, PackInfoEdit, PackPublish, PackVisibility,
	PackArchive, PackUsersView, PackUsersManage, PackSnapshotView, PackSnapshotRevert,
}

// archivedAllowed are the only permissions that can pass on an archived pack.
var archivedAllowed = map[string]struct{}{
	PackView:      {},
	PackConsume:   {},
	PackLink:      {},
	PackUsersView: {},
	PackArchive:   {}, // needed to unarchive
	// history stays readable (and cloneable) on an archived pack; revert does not pass
	PackSnapshotView: {},
}

// globalPermissions are the resource='global' permissions; they can only be
// held through a system role.
var globalPermissions = map[string]struct{}{
	PackCreate: {}, UserView: {}, UserLookup: {}, UserCreate: {},
	UserManage: {}, UserRolesAssign: {}, AuditView: {}, OidcManage: {},
}

// IsGlobal reports whether name is a resource='global' permission.
func IsGlobal(name string) bool {
	_, ok := globalPermissions[name]
	return ok
}
