package types

type PackPermission int

const (
	PackPermissionStatic PackPermission = 1
	PackPermissionView   PackPermission = 10
	PackPermissionEdit   PackPermission = 20
	// PackPermissionOwner is reserved for the pack creator. It can never be
	// granted or changed through the collaborator API.
	PackPermissionOwner PackPermission = 30
)
