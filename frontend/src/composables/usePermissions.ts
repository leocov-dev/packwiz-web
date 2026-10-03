import {computed} from "vue"
import {useAuthStore} from "@/stores/auth.ts"
import {hasPermission, type PermissionName} from "@/lib/permissions.ts"

// Global (system role) permissions of the signed-in user.
export function usePermissions() {
  const authStore = useAuthStore()
  const permissions = computed(() => authStore.user?.permissions ?? [])
  const can = (name: PermissionName) => hasPermission(permissions.value, name)

  return {permissions, can}
}
