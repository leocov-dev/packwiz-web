<script setup lang="ts">

import {ref} from 'vue';
import {useTheme} from 'vuetify';
import {usePrefStore} from "@/stores/user";

const userPrefs = usePrefStore()
const theme = useTheme();
const darkMode = ref(userPrefs.theme === 'dark')
const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');


const toggleTheme = () => {
  userPrefs.setTheme(darkMode.value ? 'light' : 'dark');
  theme.global.name.value = userPrefs.theme;
  darkMode.value = !darkMode.value;
}

const detectSystemTheme = () => {
  const prefersDark = mediaQuery.matches;
  userPrefs.setTheme(prefersDark ? 'dark' : 'light');
}

const systemThemeListener = (event: MediaQueryListEvent) => {
  if (userPrefs.theme === 'system') {
    userPrefs.setTheme(event.matches ? 'dark' : 'light');
  }
}

onMounted(() => {
  if (userPrefs.theme === 'system') {
    detectSystemTheme()
  }

  mediaQuery.addEventListener('change', systemThemeListener);
})

onUnmounted(() => {
  mediaQuery.removeEventListener('change', systemThemeListener);
})

</script>

<template>
  <v-list-item
    link
    :title="darkMode ? 'Dark' : 'Light'"
    prepend-icon="mdi-theme-light-dark"
    @click="toggleTheme"
  />
</template>
