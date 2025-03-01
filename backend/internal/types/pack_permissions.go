package types

type PackPermission int

const (
	PackPermissionStatic PackPermission = 1
	PackPermissionView   PackPermission = 10
	PackPermissionEdit   PackPermission = 20
)
