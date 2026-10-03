<script setup lang="ts">

import {useAppTheme} from "@/composables/userTheme.ts"

const {userTheme} = useAppTheme()

const options = [
  {value: "system", label: "System", icon: "mdi-laptop"},
  {value: "light", label: "Light", icon: "mdi-white-balance-sunny"},
  {value: "dark", label: "Dark", icon: "mdi-weather-night"},
]

</script>

<template>
  <div>
    <div class="text-caption text-medium-emphasis mb-1">
      Theme
    </div>
    <v-btn-toggle
      v-model="userTheme"
      mandatory
      divided
      density="comfortable"
      variant="outlined"
      color="primary"
      class="w-100"
    >
      <v-btn
        v-for="option in options"
        :key="option.value"
        :value="option.value"
        :prepend-icon="option.icon"
        :text="option.label"
        class="flex-grow-1"
        size="small"
      />
    </v-btn-toggle>
  </div>
</template>
