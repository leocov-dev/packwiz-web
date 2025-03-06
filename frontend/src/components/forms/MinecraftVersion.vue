<script setup lang="ts">
import {useCacheStore} from "@/stores/cache.ts";

const cacheStore = useCacheStore()

const version = defineModel<string | undefined>('version', {required: true})
const versions = ref<string[]>(cacheStore.minecraftVersions)


</script>

<template>
  <v-select
    v-model="version"
    label="Minecraft Version"
    :items="versions"
    hint="Minecraft server version compatible with this pack."
    persistent-hint
  />
</template>
