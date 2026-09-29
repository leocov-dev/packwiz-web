import {computed, type MaybeRefOrGetter, toValue} from "vue"
import {PackPermission, type PackResponse} from "@/interfaces/pack.ts"
import {useAuthStore} from "@/stores/auth.ts"

type PermissionSubject = Pick<PackResponse, "isArchived" | "currentUserPermission">

// Edit-level access to the pack, ignoring archived state.
export function computeHasEditAccess(pack: PermissionSubject, isAdmin: boolean): boolean {
  return pack.currentUserPermission >= PackPermission.EDIT || isAdmin
}

// Edit access on a live (non-archived) pack.
export function computeCanEdit(pack: PermissionSubject, isAdmin: boolean): boolean {
  return !pack.isArchived && computeHasEditAccess(pack, isAdmin)
}

export function usePackPermissions(pack: MaybeRefOrGetter<PermissionSubject>) {
  const authStore = useAuthStore()
  const isAdmin = computed(() => !!authStore.user?.isAdmin)

  const canEdit = computed(() => computeCanEdit(toValue(pack), isAdmin.value))
  const hasEditAccess = computed(() => computeHasEditAccess(toValue(pack), isAdmin.value))

  return {canEdit, hasEditAccess}
}
