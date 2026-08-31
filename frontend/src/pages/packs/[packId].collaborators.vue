<route lang="yaml">
meta:
  layout: app
</route>

<script setup lang="ts">
import PackCollaborators from "@/components/pack/PackCollaborators.vue";
import {useRoute} from "vue-router";
import {buildDataLoader} from "@/composables/data-loader.ts";
import type {Pack} from "@/interfaces/pack.ts";
import {fetchOnePack} from "@/services/packs.service.ts";

const route = useRoute<'/packs/[packId].collaborators'>()

const {
  isLoading,
  data: pack,
  error,
} = buildDataLoader<Pack>(async () => {
  return fetchOnePack(Number(route.params.packId), true)
})
</script>

<template>
  <div
    v-if="isLoading"
    class="ma-6"
  >
    <v-skeleton-loader
      elevation="0"
      theme="article"
      type="heading, subtitle, actions, paragraph@2"
    />
  </div>

  <v-alert
    v-else-if="error || !pack"
    class="ma-6"
    type="error"
    icon="mdi-alert"
    text="Failed to load pack."
  />

  <PackCollaborators
    v-else
    :pack="pack"
  />
</template>
