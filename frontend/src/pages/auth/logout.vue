<script setup lang="ts">
import {useAuthStore} from "@/stores/auth.ts";

const authStore = useAuthStore()

onMounted(async () => {
  await authStore.logout(false)
})
</script>

<template>
  <v-container class="fill-height">
    <v-responsive
      class="mx-auto"
      max-width="100"
    >
      <p>Logging Out...</p>
    </v-responsive>
  </v-container>
</template>
