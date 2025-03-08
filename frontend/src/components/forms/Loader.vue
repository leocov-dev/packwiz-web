<script setup lang="ts">

import {apiClient} from "@/services/api.service.ts";
import {toTitleCase} from "@/services/utils.ts";

const loader = defineModel<string | undefined>('loader', {required: true})
const version = defineModel<string>('version', {required: true})

const loaders = ref<string[]>([])
const useLatest = ref(false)

const rules = {
  loaderRequired: (value: string) => !!value || "Loader is required",
  versionRequired: (value: string) => !!value || "Loader Version is required, or select 'Use Latest'",
}

interface LoadersResponse {
  loaders: string[]
}

const getLoaderTypes = async () => {
  const response = await apiClient.get<LoadersResponse>('/v1/packwiz/loaders')
  loaders.value = response.data.loaders.map(loader => toTitleCase(loader))
}


watch(useLatest, (checked) => {
  version.value = checked ? "Latest" : ""
})


onMounted(async () => {
  await getLoaderTypes()
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
    <v-text-field
      v-model="version"
      :rules="[rules.versionRequired]"
      label="Version"
      class="me-4"
      :disabled="useLatest"
    />
    <v-checkbox
      v-model="useLatest"
      label="Use Latest"
      class="me-2"
    />
  </div>
</template>

