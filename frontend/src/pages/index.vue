<script setup lang="ts">
import { useRouter } from 'vue-router';

const router = useRouter();

onMounted(async () => {
  await router.replace('/packs');
})
</script>

<template>
  <v-container class="fill-height">
    <v-responsive
      class="mx-auto"
      max-width="100"
    >
      <p>Redirecting...</p>
    </v-responsive>
  </v-container>
</template>
