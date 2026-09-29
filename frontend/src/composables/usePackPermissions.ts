import {computed, type MaybeRefOrGetter, toValue} from "vue"
import {PackPermission, type PackResponse} from "@/interfaces/pack.ts"

type PermissionSubject = Pick<PackResponse, "isArchived" | "currentUserPermission">

// Edit-level access to the pack, ignoring archived state.
// Mirrors the backend PackPermissionGuard, which has no admin bypass.
export function computeHasEditAccess(pack: PermissionSubject): boolean {
  return pack.currentUserPermission >= PackPermission.EDIT
}

// Edit access on a live (non-archived) pack.
export function computeCanEdit(pack: PermissionSubject): boolean {
  return !pack.isArchived && computeHasEditAccess(pack)
}

export function usePackPermissions(pack: MaybeRefOrGetter<PermissionSubject>) {
  const canEdit = computed(() => computeCanEdit(toValue(pack)))
  const hasEditAccess = computed(() => computeHasEditAccess(toValue(pack)))

  return {canEdit, hasEditAccess}
}
