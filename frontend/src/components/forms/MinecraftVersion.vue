<script setup lang="ts">
import {useCacheStore} from "@/stores/cache.ts";

const cacheStore = useCacheStore()

const {includeLatest} = defineProps({includeLatest: Boolean})

let versionsList = cacheStore.minecraftVersions
if (includeLatest) {
  versionsList = [
    // TODO: request including these formatted strings is not valid
    `Latest (${cacheStore.minecraftLatest})`,
    `Latest Snapshot (${cacheStore.minecraftSnapshot})`,
    ...versionsList
  ]
}

const version = defineModel<string | undefined>('version', {required: true})

if (!!version.value && !versionsList.includes(version.value)) {
  versionsList = [
    version.value,
    ...versionsList,
  ]
}

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
