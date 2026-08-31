import {useAuthStore} from "@/stores/auth";
import type {NavigationGuardNext, RouteLocationNormalizedGeneric, RouteLocationNormalizedLoaded} from "vue-router";

export async function authGuard(
  to: RouteLocationNormalizedGeneric,
  from: RouteLocationNormalizedLoaded,
  next: NavigationGuardNext,
): Promise<void> {
  const authStore = useAuthStore()
  await authStore.checkAuth(false)

  if (to.meta.noAuth || authStore.isAuthenticated) {
    next()
  } else {
    next({
        path: '/auth/login',
        query: { redirect: to.fullPath },
      }
    )
  }
}
