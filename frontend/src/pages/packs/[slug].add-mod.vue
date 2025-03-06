<route lang="yaml">
meta:
  layout: app
</route>

<script lang="ts" setup>
import {useRoute} from "vue-router";
import {buildDataLoader} from "@/composables/data-loader.ts";
import type {Pack} from "@/interfaces/pack.ts";
import {fetchOnePack} from "@/services/packs.service.ts";
import PackAddModForm from "@/components/PackAddModForm.vue";

const route = useRoute<'/packs/[slug].add-mod'>()

const {
  isLoading,
  data: pack,
} = buildDataLoader<Pack>(async () => {
  return fetchOnePack(route.params.slug)
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

  <PackAddModForm
    v-else-if="pack"
    :pack="pack"
  />
</template>
