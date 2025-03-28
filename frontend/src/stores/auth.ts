import {defineStore} from 'pinia'
import {
  changePassword,
  getCurrentUser,
  userLogin,
  userLogout,
  invalidateCurrentUserSessions
} from "@/services/user.service"
import {User} from "@/interfaces/user"
import router from "@/router";
import type {RouteLocationRaw} from "vue-router";
import {usePrefStore} from "@/stores/user";
import {initializeCacheStore} from "@/stores/cache.ts";
import {AxiosError} from "axios";


interface AuthState {
  user: User | null;
}

interface AuthActions {
  checkAuth(force: boolean): Promise<void>;

  refreshUser(): Promise<void>;

  login(username: string, password: string): Promise<void>;

  logout(redirect: boolean): Promise<void>;

  changePassword(oldPassword: string, newPassword: string): Promise<string>;

  invalidateSessions(): Promise<void>;
}

interface AuthGetters {
  isAuthenticated(state: AuthState): boolean;

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  [key: string]: any;
}

export const useAuthStore = defineStore<'auth', AuthState, AuthGetters, AuthActions>('auth', {
  state: () => ({
    user: null
  }),
  getters: {
    isAuthenticated(state: AuthState) {
      return state.user !== null
    },
  },
  actions: {
    async checkAuth(force = false) {
      if (!force && this.isAuthenticated) {
        console.debug('User is cached')
        return
      }

      try {
        console.debug('Checking user...')
        this.user = await getCurrentUser()

        await initializeCacheStore()

        const userPrefs = usePrefStore()
        userPrefs.loadPreferences(this.user.id)
      } catch {
        this.user = null
      }
    },
    async refreshUser() {
      if (!this.isAuthenticated) return

      this.user = await getCurrentUser()
      const userPrefs = usePrefStore()
      userPrefs.loadPreferences(this.user.id)
    },
    async login(username: string, password: string) {
      await userLogin(username, password)
      await this.checkAuth(true)
    },
    async logout(redirect = true) {
      await userLogout()
      this.user = null

      const params: RouteLocationRaw = {path: '/auth/login'}
      if (redirect) {
        params['query'] = {redirect: router.currentRoute.value.fullPath || '/'}
      }

      await router.push(params)
    },
    async changePassword(oldPassword: string, newPassword: string) {
      try {
        await changePassword(oldPassword, newPassword)
        await this.logout(true)
        return ""
      } catch (e) {
        if (e instanceof AxiosError) {
          return e.response?.data?.msg || "Unknown error..."
        }
        return "Unknown error..."
      }
    },
    async invalidateSessions() {
      await invalidateCurrentUserSessions()
    },
  },
})
