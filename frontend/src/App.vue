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
    location="top"
  >
    {{ snackbar.message }}
    <template #actions>
      <v-btn
        icon="mdi-close"
        @click="snackbar.closeSnackbar"
      />
    </template>
  </v-snackbar>
</template>
