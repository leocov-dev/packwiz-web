<route lang="yaml">
meta:
  layout: app
</route>

<script setup lang="ts">
import {useRoute} from "vue-router";
import {buildDataLoader} from "@/composables/data-loader.ts";
import {type Pack, type Mod} from "@/interfaces/pack.ts";
import {fetchOnePack} from "@/services/packs.service.ts";
import {fetchOneMod} from "@/services/mods.service.ts";

const route = useRoute<'/packs/[slug].mod.[name]'>()

const {
  isLoading,
  data,
} = buildDataLoader<{ pack: Pack, mod: Mod }>(async () => {
  const pack = await fetchOnePack(route.params.slug, true)
  const mod = await fetchOneMod(route.params.slug, route.params.name)
  return {pack, mod}
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

  <EditModForm
    v-else-if="data?.pack && data?.mod"
    :pack="data.pack"
    :mod="data.mod"
  />
</template>

