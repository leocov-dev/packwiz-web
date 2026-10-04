<script setup lang="ts">
import {SOURCE_REPO_URL} from "@/lib/links.ts";
</script>

<template>
  <v-footer class="justify-center flex-grow-0 py-2">
    <a
      :href="SOURCE_REPO_URL"
      target="_blank"
      rel="noopener noreferrer"
      class="d-inline-flex align-center ga-1 text-caption text-medium-emphasis"
    >
      <v-icon
        icon="mdi-github"
        size="small"
      />
      Packwiz Web on GitHub
    </a>
  </v-footer>
</template>
