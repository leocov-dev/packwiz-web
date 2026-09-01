<script setup lang="ts">
import type {Mod} from "@/interfaces/pack.ts";
import {filterModsBySide, type ModSide} from "@/lib/mod-filters.ts";

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
          v-if="canEdit"
          class="me-3"
          color="primary"
          variant="flat"
          prepend-icon="mdi-plus"
          text="Add Mod"
          @click="$emit('add-mod')"
        />
      </v-toolbar>
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
