import {type MaybeRefOrGetter, toValue} from "vue"
import type {PackResponse} from "@/interfaces/pack.ts"
import {hasPermission, type PermissionName} from "@/lib/permissions.ts"

type PermissionSubject = Pick<PackResponse, "permissions">

// Effective permissions on a pack, as computed by the backend (system and pack
// roles merged, archived-pack rules applied).
export function computeCan(pack: PermissionSubject, name: PermissionName): boolean {
  return hasPermission(pack.permissions, name)
}

export function usePackPermissions(pack: MaybeRefOrGetter<PermissionSubject>) {
  const can = (name: PermissionName) => computeCan(toValue(pack), name)

  return {can}
}
