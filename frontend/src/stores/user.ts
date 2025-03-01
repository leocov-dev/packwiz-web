import {defineStore} from 'pinia'
import {useTheme} from "vuetify";

export const usePrefStore = defineStore('userPreferences', {
  state: () => ({
    theme: 'dark' as 'dark' | 'light' | 'system',

    $userId: null as number | null,
  }),

  actions: {
    loadPreferences(userId: number) {
      this.$userId = userId
      const savedPrefs = localStorage.getItem(`user-prefs-${userId}`)

      if (savedPrefs) {
        const parsed = JSON.parse(savedPrefs)
        this.$patch(parsed)
      }
    },
    savePreferences() {
      if (!this.$userId) return
      localStorage.setItem(`user-prefs-${this.$userId}`, JSON.stringify(this.$state))
    },
    clearPreferences() {
      if (!this.$userId) return
      localStorage.removeItem(`user-prefs-${this.$userId}`)
      this.$reset()
    },
    setTheme(theme: 'dark' | 'light' | 'system') {
      this.theme = theme
      this.savePreferences()
    }
  }
})
