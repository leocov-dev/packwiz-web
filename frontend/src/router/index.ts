import {createRouter, createWebHistory} from 'vue-router'
import {routes} from 'vue-router/auto-routes'
// @ts-expect-error [TS2307]
import {setupLayouts} from 'virtual:vue-layouts'
import {authGuard} from "@/router/auth-guard.ts";
import type {RouteLocationNormalizedGeneric} from "vue-router";


const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: setupLayouts(routes),
})

router.beforeEach(authGuard)


// Workaround for https://github.com/vitejs/vite/issues/11804
// eslint-disable-next-line @typescript-eslint/no-explicit-any
router.onError((err: any, to: RouteLocationNormalizedGeneric) => {
  if (err?.message?.includes?.('Failed to fetch dynamically imported module')) {
    if (!localStorage.getItem('vuetify:dynamic-reload')) {
      console.log('Reloading page to fix dynamic import error')
      localStorage.setItem('vuetify:dynamic-reload', 'true')
      location.assign(to.fullPath)
    } else {
      console.error('Dynamic import error, reloading page did not fix it', err)
    }
  } else {
    console.error(err)
  }
})

router.isReady().then(() => {
  localStorage.removeItem('vuetify:dynamic-reload')
})

export default router
