<script setup lang="ts">
import type {Mod} from "@/interfaces/pack.ts";
import {
  applyModListState,
  buildDependentsMap,
  buildModListQuery,
  countMods,
  dependentNames as getDependentNames,
  findOrphanedDependencies,
  formatFilteredCount,
  formatModCounts,
  hasActiveModFilters,
  parseModListQuery,
  type ModShow,
  type ModSide,
  type ModSort,
} from "@/lib/mod-filters.ts";

const {packId, mods, canEdit} = defineProps<{
  packId: number,
  mods: Mod[],
  canEdit: boolean,
}>()

defineEmits(['add-mod', 'reload'])

const route = useRoute()
const router = useRouter()

// Filter state lives in the route query: a reload remounts this component
// (parent swaps in a skeleton), so local refs alone would be lost.
const initial = parseModListQuery(route.query)
const search = ref<string>(initial.q)
const sort = ref<ModSort>(initial.sort)
const sideFilter = ref<ModSide>(initial.side)
const show = ref<ModShow[]>(initial.show)
const currentPage = ref(1)

const listState = computed(() => ({
  q: search.value ?? '',
  sort: sort.value,
  side: sideFilter.value,
  show: show.value,
}))

watch(listState, (state) => {
  currentPage.value = 1
  router.replace({query: {...route.query, q: undefined, sort: undefined, side: undefined, show: undefined, ...buildModListQuery(state)}})
})

const sortOptions = [
  {title: 'Name (A-Z)', value: 'name'},
  {title: 'Recently updated', value: 'updated'},
]
const sideOptions = [
  {title: 'All Sides', value: ''},
  {title: 'Client', value: 'client'},
  {title: 'Server', value: 'server'},
  {title: 'Client + Server', value: 'both'},
]

const sortedMods = computed(() => applyModListState(mods, listState.value))

const dependents = computed(() => buildDependentsMap(mods))
const dependentNamesFor = (mod: Mod) => getDependentNames(mod, dependents.value)
const orphanNamesFor = (mod: Mod) => findOrphanedDependencies(mod, mods, dependents.value).map(m => m.name)

const counts = computed(() => countMods(mods))
const hasActiveFilters = computed(() => hasActiveModFilters(listState.value))
const countLabel = computed(() => hasActiveFilters.value
  ? formatFilteredCount(sortedMods.value.length, counts.value.total)
  : formatModCounts(counts.value))

const hasMods = computed(() => mods.length > 0)

const clearFilters = () => {
  search.value = ''
  sideFilter.value = ''
  show.value = []
}

const isFirstDependency = (mod: Mod, items: readonly Mod[], index: number) => {
  if (!mod.isDependency) return false;
  return index === 0 || !items[index - 1].isDependency;
};

</script>

<template>
  <v-data-iterator
    v-model:page="currentPage"
    :items="sortedMods"
    items-per-page="20"
  >
    <template #header>
      <v-toolbar height="auto">
        <v-toolbar-title>Mods</v-toolbar-title>
        <span class="text-body-2 text-medium-emphasis me-3 d-none d-sm-inline">{{ countLabel }}</span>
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
          :items="sideOptions"
          max-width="180"
          class="me-3"
          density="compact"
          variant="solo"
          hide-details
        />
        <v-select
          v-model="sort"
          :items="sortOptions"
          max-width="200"
          class="me-3"
          density="compact"
          variant="solo"
          label="Sort"
          hide-details
        />
      </v-toolbar>
      <v-chip-group
        v-model="show"
        multiple
        class="px-4 py-2"
        aria-label="Show only"
      >
        <v-chip
          value="pinned"
          filter
          variant="outlined"
          text="Pinned"
        />
        <v-chip
          value="optional"
          filter
          variant="outlined"
          text="Optional"
        />
        <v-chip
          value="dependencies"
          filter
          variant="outlined"
          text="Dependencies"
        />
      </v-chip-group>
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
            :dependent-names="dependentNamesFor(item.raw)"
            :orphan-names="orphanNamesFor(item.raw)"
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
