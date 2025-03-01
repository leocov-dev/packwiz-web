<script setup lang="ts">

const model = defineModel({required: true, type: String})

const {density} = defineProps<{density?: 'default' | 'comfortable' | 'compact' }>()

const rawSearch = ref(model.value)

const onSearch = () => {
  model.value = rawSearch.value
}
</script>

<template>
  <v-text-field
    v-model="rawSearch"
    :density="density || 'comfortable'"
    placeholder="Search"
    append-inner-icon="mdi-magnify"
    variant="solo"
    clearable
    hide-details
    @click:append-inner="onSearch"
    @click:clear="onSearch"
    @keyup.enter="onSearch"
  />
</template>
