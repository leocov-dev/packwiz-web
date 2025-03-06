import {defineStore} from 'pinia'
import {getCurrentUser, userLogin, userLogout} from "@/services/user.service"
import {User} from "@/interfaces/user"
import router from "@/router";
import type {RouteLocationRaw} from "vue-router";
import {usePrefStore} from "@/stores/user";


interface AuthState {
  user: User | null;
}

interface AuthActions {
  checkUser(force: boolean): Promise<void>;

  login(username: string, password: string): Promise<void>;

  logout(redirect: boolean): Promise<void>;
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
    async checkUser(force = false) {
      if (!force && this.isAuthenticated) {
        console.debug('User is cached')
        return
      }

      try {
        console.debug('Checking user...')
        this.user = await getCurrentUser()

        const userPrefs = usePrefStore()
        userPrefs.loadPreferences(this.user.id)
      } catch {
        this.user = null
      }
      console.debug('User is ', this.isAuthenticated ? 'authenticated' : 'not authenticated')
    },
    async login(username: string, password: string) {
      await userLogin(username, password)
      await this.checkUser(true)
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
  },
})
