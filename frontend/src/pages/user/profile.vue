<route lang="yaml">
meta:
  layout: app
</route>

<script setup lang="ts">
import {useAuthStore} from "@/stores/auth.ts";

const authStore = useAuthStore()

const user = authStore.user
</script>

<template>
  <v-container class="ml-8 mr-8 mt-3">
    <h1>Profile</h1>

    <pre>{{ user }}</pre>
  </v-container>
</template>

