import {ref, onMounted, onUnmounted, watch} from 'vue'
import {useTheme} from 'vuetify'
import {usePrefStore, type ThemePref} from '@/stores/user'

export function useAppTheme() {
  const userPrefs = usePrefStore()
  const theme = useTheme()
  const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)')
  const userTheme = ref<ThemePref>(userPrefs.theme)

  const getTrueTheme = () => {
    if (userTheme.value === 'system') {
      return mediaQuery.matches ? 'dark' : 'light'
    }
    return userTheme.value
  }

  const updateAppTheme = () => {
    userPrefs.setTheme(userTheme.value)
    theme.global.name.value = getTrueTheme()
  }

  const systemThemeListener = (event: MediaQueryListEvent) => {
    if (userPrefs.theme === 'system') {
      theme.global.name.value = event.matches ? 'dark' : 'light'
    }
  }

  watch(userTheme, () => {
    updateAppTheme()
  })

  onMounted(() => {
    if (userPrefs.theme === 'system') {
      userPrefs.setTheme('system')
    }

    updateAppTheme()
    mediaQuery.addEventListener('change', systemThemeListener)
  })

  onUnmounted(() => {
    mediaQuery.removeEventListener('change', systemThemeListener)
  })

  return {
    userTheme,
    updateAppTheme
  }
}
