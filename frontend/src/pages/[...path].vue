<route lang="yaml">
meta:
  noAuth: true
</route>

<script setup lang="ts">
</script>

<template>
  <v-main class="fill-height">
    <v-responsive
      class="d-flex justify-center align-center fill-height mx-auto"
      max-width="300"
    >
      <v-img
        class="mb-4"
        height="100"
        src="@/assets/logo.png"
      />

      <div class="mb-5 text-center">
        <h1 class="font-weight-bold mb-5">
          Page Not Found
        </h1>

        <router-link to="/">
          Return Home
        </router-link>
      </div>
    </v-responsive>
  </v-main>
</template>

