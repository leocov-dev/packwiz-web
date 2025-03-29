<script lang="ts">
export interface Filters {
  [key: string]: boolean
}

export interface FiltersConfig {
  [key: keyof Filters]: {
    title: string,
  }
}

export const reduceFilters = (filters: Record<string, boolean>) => {
  return Object.keys(filters).filter(key => filters[key])
}
</script>

<script setup lang="ts">

const model = defineModel({required: true, type: Object as () => Filters})

const {config} = defineProps<{
  config: FiltersConfig
}>()

</script>

<template>
  <v-menu
    :close-on-content-click="false"
    location="bottom end"
  >
    <template #activator="{ props }">
      <v-btn
        icon="mdi-filter-menu-outline"
        v-bind="props"
      />
    </template>

    <v-card min-width="200">
      <v-list v-if="model">
        <v-list-item
          v-for="(val, key) in config"
          :key="key"
        >
          <v-checkbox
            v-model="model[key]"
            :label="val.title"
            :hide-details="true"
            density="comfortable"
          />
        </v-list-item>
      </v-list>
    </v-card>
  </v-menu>
</template>
