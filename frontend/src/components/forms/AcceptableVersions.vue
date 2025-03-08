<script setup lang="ts">
import {useCacheStore} from "@/stores/cache.ts";

const cacheStore = useCacheStore()

const acceptableVersions = defineModel<string[]>('acceptableVersions', {required: true})

const versions = ref(cacheStore.minecraftVersions)
</script>

<template>
  <v-select
    v-model="acceptableVersions"
    :items="versions"
    hint="Client Minecraft versions that are compatible with this pack. Leave blank to allow only matching server version."
    label="Client Acceptable Versions"
    chips
    multiple
    persistent-hint
  />
</template>

