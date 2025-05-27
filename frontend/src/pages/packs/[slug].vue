<route lang="yaml">
meta:
  layout: app
</route>

<script lang="ts" setup>

import {useRoute} from "vue-router";
import PackDetails from "@/components/pack/PackDetails.vue";
import {buildDataLoader} from "@/composables/data-loader.ts";
import type {PackResponse} from "@/interfaces/pack.ts";
import {fetchOnePack} from "@/services/packs.service.ts";

const route = useRoute<'/packs/[slug]'>()

const {
  isLoading,
  data: pack,
  reload,
} = buildDataLoader<PackResponse>(async () => {
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

  <PackDetails
    v-else-if="pack"
    :pack="pack"
    @reload="reload"
  />
</template>
