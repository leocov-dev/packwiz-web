import {createRouter, createWebHistory} from 'vue-router/auto'
import {routes} from 'vue-router/auto-routes'
// @ts-ignore [TS2307]
import {setupLayouts} from 'virtual:vue-layouts'
import {useAuthStore} from "@/stores/auth";


const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: setupLayouts(routes),
})

router.beforeEach(async (to, from, next) => {
  const authStore = useAuthStore()
  await authStore.checkUser(false)

  if (to.meta.noAuth || authStore.isAuthenticated) {
    console.debug('route guard auth check passed')
    next()
  } else {
    console.debug('route guard auth check failed')
    next({
        path: '/auth/login',
        query: { redirect: to.fullPath },
      }
    )
  }

})


// Workaround for https://github.com/vitejs/vite/issues/11804
router.onError((err, to) => {
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
