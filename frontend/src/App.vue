<script lang="ts" setup>
import { useSnackbarStore } from '@/stores/snackbar';

import { useAppTheme } from '@/composables/userTheme';

const {  } = useAppTheme()



const snackbar = useSnackbarStore();
</script>

<template>
  <router-view />

  <v-snackbar
    v-model="snackbar.show"
    :timeout="snackbar.timeout"
    :color="snackbar.color"
    location="bottom"
  >
    {{ snackbar.message }}
    <template #actions>
      <v-btn
        v-if="snackbar.link"
        :text="snackbar.link.text"
        :to="snackbar.link.to"
        variant="text"
        @click="snackbar.closeSnackbar"
      />
      <v-btn
        icon="mdi-close"
        aria-label="Close"
        @click="snackbar.closeSnackbar"
      />
    </template>
  </v-snackbar>
</template>
