<route lang="yaml">
meta:
  layout: app
</route>

<script setup lang="ts">
import PackEditForm from "@/components/pack/PackEditForm.vue";
import {useRoute} from "vue-router";
import {buildDataLoader} from "@/composables/data-loader.ts";
import type {Pack} from "@/interfaces/pack.ts";
import {fetchOnePack} from "@/services/packs.service.ts";

const route = useRoute<'/packs/[packId].add-mod'>()

const {
  isLoading,
  data: pack,
} = buildDataLoader<Pack>(async () => {
  return fetchOnePack(route.params.packId)
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

  <PackEditForm
    v-else-if="pack"
    :pack="pack"
  />
</template>
