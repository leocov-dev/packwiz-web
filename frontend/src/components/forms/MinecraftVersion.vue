<script setup lang="ts">
import {useCacheStore} from "@/stores/cache.ts";

const cacheStore = useCacheStore()

const {includeLatest} = defineProps({includeLatest: Boolean})

let versionsList = cacheStore.minecraftVersions
if (includeLatest) {
  versionsList = [
    "Latest",
    "Latest Snapshot",
    ...versionsList
  ]
}

const version = defineModel<string | undefined>('version', {required: true})
const versions = ref<string[]>(versionsList)

const rules = {
  versionRequired: (value: string) => !!value || "Minecraft Version is required",
}

</script>

<template>
  <v-select
    v-model="version"
    :rules="[rules.versionRequired]"
    :items="versions"
    label="Minecraft Version"
    hint="Minecraft server version compatible with this pack."
    persistent-hint
    persistent-placeholder
  />
</template>
