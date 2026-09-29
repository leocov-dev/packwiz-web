<script setup lang="ts">
import type {Mod} from "@/interfaces/pack.ts";
import {countMods, filterModsBySide, type ModSide} from "@/lib/mod-filters.ts";

const {packId, mods, canEdit} = defineProps<{
  packId: number,
  mods: Mod[],
  canEdit: boolean,
}>()

defineEmits(['add-mod', 'reload'])


const search = ref<string>('')
const sideFilter = ref<ModSide>('')

const sortedMods = computed(() => {
  const filteredMods = filterModsBySide(mods, sideFilter.value)

  const regularMods = filteredMods.filter(mod => !mod.isDependency)
    .sort((a, b) => a.name.localeCompare(b.name));

  const dependencyMods = filteredMods.filter(mod => mod.isDependency)
    .sort((a, b) => a.name.localeCompare(b.name));

  return [...regularMods, ...dependencyMods];
})

const counts = computed(() => countMods(mods))
const countLabel = computed(() => `${counts.value.total} mods · ${counts.value.dependencies} dependencies`)

const hasMods = computed(() => mods.length > 0)
const hasActiveFilters = computed(() => !!search.value || !!sideFilter.value)

const clearFilters = () => {
  search.value = ''
  sideFilter.value = ''
}

const isFirstDependency = (mod: Mod, items: readonly Mod[], index: number) => {
  if (!mod.isDependency) return false;
  return index === 0 || !items[index - 1].isDependency;
};

</script>

<template>
  <v-data-iterator
    :items="sortedMods"
    :search="search"
    items-per-page="20"
  >
    <template #header>
      <v-toolbar class="d-flex flex-wrap">
        <v-toolbar-title>Mods</v-toolbar-title>
        <span class="text-body-2 text-medium-emphasis me-3">{{ countLabel }}</span>
        <v-text-field
          v-model="search"
          max-width="300"
          class="me-3"
          density="compact"
          placeholder="Search"
          prepend-inner-icon="mdi-magnify"
          variant="solo"
          clearable
          hide-details
        />
        <v-select
          v-model="sideFilter"
          :items="[
            {title: 'All Sides', value: ''},
            {title: 'Client', value: 'client'},
            {title: 'Server', value: 'server'},
            {title: 'Client + Server', value: 'both'},
          ]"
          max-width="180"
          class="me-3"
          density="compact"
          variant="solo"
          hide-details
        />
        <v-btn
          v-if="canEdit && hasMods"
          class="me-3"
          color="primary"
          variant="flat"
          prepend-icon="mdi-plus"
          text="Add Mod"
          @click="$emit('add-mod')"
        />
      </v-toolbar>
    </template>

    <template #no-data>
      <div
        v-if="!hasMods"
        class="d-flex flex-column align-center text-center pa-10"
      >
        <v-icon
          icon="mdi-package-variant"
          size="48"
          class="mb-3 text-medium-emphasis"
        />
        <div class="text-h6">
          No mods yet
        </div>
        <div
          v-if="canEdit"
          class="text-body-2 text-medium-emphasis mb-4"
        >
          Add your first mod to get started.
        </div>
        <v-btn
          v-if="canEdit"
          color="primary"
          variant="flat"
          prepend-icon="mdi-plus"
          text="Add Mod"
          @click="$emit('add-mod')"
        />
      </div>
      <div
        v-else
        class="d-flex flex-column align-center text-center pa-10"
      >
        <v-icon
          icon="mdi-filter-off-outline"
          size="48"
          class="mb-3 text-medium-emphasis"
        />
        <div class="text-h6 mb-4">
          No mods match your filters
        </div>
        <v-btn
          v-if="hasActiveFilters"
          variant="tonal"
          text="Clear filters"
          @click="clearFilters"
        />
      </div>
    </template>

    <template #default="{items}">
      <v-list>
        <v-list-item
          v-for="(item, index) in items"
          :key="item.raw.id"
          :class="{'first-dependency': isFirstDependency(item.raw, items.map(i => i.raw), index)}"
        >
          <ModCard
            :pack-id="packId"
            :mod="item.raw"
            :can-edit="canEdit"
            @reload="$emit('reload')"
          />
        </v-list-item>
      </v-list>
    </template>

    <template #footer="{ page, pageCount, prevPage, nextPage }">
      <div
        v-if="pageCount > 1"
        class="d-flex align-center justify-center pa-4"
      >
        <v-btn
          icon="mdi-chevron-left"
          variant="text"
          density="comfortable"
          :disabled="page === 1"
          @click="prevPage"
        />
        <span class="mx-4">Page {{ page }} of {{ pageCount }}</span>
        <v-btn
          icon="mdi-chevron-right"
          variant="text"
          density="comfortable"
          :disabled="page === pageCount"
          @click="nextPage"
        />
      </div>
    </template>
  </v-data-iterator>
</template>

<style scoped>
.first-dependency {
  margin-top: 24px !important;
}
</style>
