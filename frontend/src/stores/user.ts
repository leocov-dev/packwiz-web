import {defineStore} from 'pinia'


export type ThemePref = 'dark' | 'light' | 'system'

export const usePrefStore = defineStore('userPreferences', {
  state: () => ({
    theme: 'dark' as ThemePref,

    $userId: null as number | null,
  }),

  actions: {
    loadPreferences(userId: number) {
      this.$userId = userId
      const savedPrefs = localStorage.getItem(`user-prefs-${userId}`)

      if (savedPrefs) {
        try {
          const parsed = JSON.parse(savedPrefs)
          this.$patch(parsed)
        } catch (e) {
          console.warn(`Corrupted preferences for user ${userId}, resetting`, e)
          localStorage.removeItem(`user-prefs-${userId}`)
        }
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
    setTheme(theme: ThemePref) {
      this.theme = theme
      this.savePreferences()
    }
  }
})
