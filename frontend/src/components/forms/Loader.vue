<script setup lang="ts">

import {useCacheStore, type LoaderVersions} from "@/stores/cache.ts";

const cacheStore = useCacheStore()

const {minecraftVersion} = defineProps({minecraftVersion: {type: String, default: () => ""}})

const loader = defineModel<keyof LoaderVersions | undefined>('loader', {required: true})
const version = defineModel<string>('version', {default: ""})

const loaderVersions = ref<string[]>([])
const loaders = ref<string[]>(cacheStore.loaders)
const noDataText = ref("")

const rules = {
  loaderRequired: (value: string) => !!value || "Loader is required",
  versionRequired: (value: string) => !!value || "Loader Version is required",
}

// ----
const getLoaderVersions = (loader: keyof LoaderVersions | undefined, mcVersion: string) => {
  if (!loader) {
    noDataText.value = "Must Select Loader Type First!"
    return []
  }

  loader = loader.toLowerCase() as keyof LoaderVersions

  const listOrMap = cacheStore.loaderVersions[loader]

  if (Array.isArray(listOrMap)) {
    return listOrMap
  }

  if (!mcVersion) {
    noDataText.value = "Must Select Minecraft Version First!"
    return []
  }

  if (mcVersion.toLowerCase().includes('snapshot')) {
    mcVersion = cacheStore.minecraftSnapshot.split('-')[0]
  } else if (mcVersion.toLowerCase().includes('latest')) {
    mcVersion = cacheStore.minecraftLatest
  }

  noDataText.value = `No versions available for Minecraft ${mcVersion}`
  return listOrMap[mcVersion.split('-')[0]]
}

watch(() => minecraftVersion, (newVersion) => {
  version.value = ""
  loaderVersions.value = getLoaderVersions(loader.value, newVersion)
})

watch(loader, (newLoader) => {
  version.value = ""
  loaderVersions.value = getLoaderVersions(newLoader, minecraftVersion)
})

onMounted(() => {
  loaderVersions.value = getLoaderVersions(loader.value, minecraftVersion)
})

</script>

<template>
  <div class="d-flex">
    <v-select
      v-model="loader"
      :items="loaders"
      :rules="[rules.loaderRequired]"
      label="Loader"
      class="me-6"
    />
    <v-select
      v-model="version"
      :items="loaderVersions"
      :rules="[rules.versionRequired]"
      :no-data-text="noDataText"
      label="Version"
      class="me-4"
    />
  </div>
</template>

