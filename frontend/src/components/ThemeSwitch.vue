<script setup lang="ts">

import {useAppTheme} from "@/composables/userTheme.ts"

const {userTheme} = useAppTheme()

</script>

<template>
  <v-radio-group
    v-model="userTheme"
    inline
    class="d-flex"
  >
    <v-radio value="system">
      <template #label>
        <v-icon icon="mdi-desktop-classic" />
      </template>
    </v-radio>
    <v-radio value="dark">
      <template #label>
        <v-icon icon="mdi-brightness-3" />
      </template>
    </v-radio>
    <v-radio value="light">
      <template #label>
        <v-icon icon="mdi-brightness-5" />
      </template>
    </v-radio>
  </v-radio-group>
</template>
