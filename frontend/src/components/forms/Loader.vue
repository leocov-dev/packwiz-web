<script setup lang="ts">

import {apiClient} from "@/services/api.service.ts";

const loader = defineModel<string | undefined>('loader', {required: true})
const version = defineModel<string>('version', {required: true})

const loaders = ref<string[]>([])
const useLatest = ref(false)

const getLoaderTypes = async () => {
  const response = await apiClient.get('/v1/packwiz/loaders')
  loaders.value = response.data.loaders
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
      label="Loader"
      :items="loaders"
      class="me-6"
    />
    <v-text-field
      v-model="version"
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

