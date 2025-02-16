package types

type PackPermission string

const (
	PackPermissionStatic PackPermission = "static"
	PackPermissionView   PackPermission = "view"
	PackPermissionAdmin  PackPermission = "admin"
)
