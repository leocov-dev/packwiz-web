<script setup lang="ts">
import type {Mod} from "@/interfaces/pack.ts";
import {updateAvailableIds, type UpdateChecksMap} from "@/lib/update-checks.ts";
import {
  applyModListState,
  applyPinOverrides,
  buildDependentNamesMap,
  buildDependentsMap,
  buildOrphanNamesMap,
  buildModListQuery,
  countMods,
  formatFilteredCount,
  formatModCounts,
  hasActiveModFilters,
  parseModListQuery,
  type ModShow,
  type ModSide,
  type ModSort,
} from "@/lib/mod-filters.ts";

const {packId, mods, canEdit, updateChecks = new Map()} = defineProps<{
  packId: number,
  mods: Mod[],
  canEdit: boolean,
  updateChecks?: UpdateChecksMap,
}>()

const emit = defineEmits<{
  (e: 'add-mod'): void
  (e: 'reload'): void
  (e: 'pin-overrides', overrides: ReadonlyMap<number, boolean>): void
}>()

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

// The search field stays instantly responsive; the applied value (used for
// filtering, paging and the route query) trails it by a short debounce.
const SEARCH_DEBOUNCE_MS = 250
const appliedSearch = ref<string>(initial.q)
let searchTimer: ReturnType<typeof setTimeout> | undefined
watch(search, (value) => {
  clearTimeout(searchTimer)
  if (!value) {
    appliedSearch.value = ''
    return
  }
  searchTimer = setTimeout(() => {
    appliedSearch.value = value
  }, SEARCH_DEBOUNCE_MS)
})
onBeforeUnmount(() => clearTimeout(searchTimer))

const listState = computed(() => ({
  q: appliedSearch.value ?? '',
  sort: sort.value,
  side: sideFilter.value,
  show: show.value,
}))

watch(listState, (state) => {
  currentPage.value = 1
  router.replace({query: {...route.query, q: undefined, sort: undefined, side: undefined, show: undefined, ...buildModListQuery(state)}})
})

// Successful pin/unpin results not yet reflected in the `mods` prop. Applying
// them here keeps cards, filters, counts and sort consistent without a reload
// (which would remount the list and reset page/scroll).
const pinOverrides = ref(new Map<number, boolean>())
watch(() => mods, () => {
  pinOverrides.value = new Map()
  emit('pin-overrides', pinOverrides.value)
})
const onPinned = (id: number, value: boolean) => {
  const next = new Map(pinOverrides.value)
  next.set(id, value)
  pinOverrides.value = next
  emit('pin-overrides', next)
}
const effectiveMods = computed(() => applyPinOverrides(mods, pinOverrides.value))

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

const updateIds = computed(() => updateAvailableIds(updateChecks))
const sortedMods = computed(() => applyModListState(effectiveMods.value, listState.value, updateIds.value))

const dependents = computed(() => buildDependentsMap(effectiveMods.value))
const dependentNamesById = computed(() => buildDependentNamesMap(dependents.value))
const orphanNamesById = computed(() => buildOrphanNamesMap(effectiveMods.value, dependents.value))
const NO_NAMES: string[] = []
const dependentNamesFor = (mod: Mod) => dependentNamesById.value.get(mod.id) ?? NO_NAMES
const orphanNamesFor = (mod: Mod) => orphanNamesById.value.get(mod.id) ?? NO_NAMES

const counts = computed(() => countMods(effectiveMods.value))
const hasActiveFilters = computed(() => hasActiveModFilters(listState.value))
const countLabel = computed(() => hasActiveFilters.value
  ? formatFilteredCount(sortedMods.value.length, counts.value.total)
  : formatModCounts(counts.value))

const hasMods = computed(() => mods.length > 0)
// "Updates available" selected but nothing was ever checked (or results were reset)
const needsCheckHint = computed(() => show.value.includes('updates') && updateChecks.size === 0)

const clearFilters = () => {
  search.value = ''
  appliedSearch.value = ''
  sideFilter.value = ''
  show.value = []
}

// Gap only where a dependency directly follows a regular mod on the same page.
const isFirstDependency = (items: readonly {raw: Mod}[], index: number) =>
  index > 0 && items[index].raw.isDependency && !items[index - 1].raw.isDependency

</script>

<template>
  <v-data-iterator
    v-model:page="currentPage"
    :items="sortedMods"
    items-per-page="20"
  >
    <template #header>
      <v-toolbar height="auto">
        <div class="d-flex flex-wrap align-center ga-2 pa-2 w-100">
          <div class="d-flex flex-column flex-sm-row align-sm-center me-sm-3">
            <v-toolbar-title class="flex-grow-0">
              Mods
            </v-toolbar-title>
            <span class="text-body-2 text-medium-emphasis ms-4 ms-sm-2">{{ countLabel }}</span>
          </div>
          <v-text-field
            v-model="search"
            max-width="300"
            class="flex-grow-1"
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
            min-width="140"
            class="flex-grow-1"
            density="compact"
            variant="solo"
            hide-details
          />
          <v-select
            v-model="sort"
            :items="sortOptions"
            max-width="200"
            min-width="140"
            class="flex-grow-1"
            density="compact"
            variant="solo"
            label="Sort"
            hide-details
          />
        </div>
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
          label
          variant="tonal"
          text="Pinned"
        />
        <v-chip
          value="optional"
          filter
          label
          variant="tonal"
          text="Optional"
        />
        <v-chip
          value="dependencies"
          filter
          label
          variant="tonal"
          text="Dependencies"
        />
        <v-chip
          value="updates"
          filter
          label
          variant="tonal"
          text="Updates available"
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
        <div
          v-if="needsCheckHint"
          class="text-body-2 text-medium-emphasis mb-4"
        >
          {{ canEdit ? "No update check results yet. Run 'Check for updates' first." : "No update check results yet." }}
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
          :class="{'mt-6': isFirstDependency(items, index)}"
        >
          <ModCard
            :pack-id="packId"
            :mod="item.raw"
            :can-edit="canEdit"
            :dependent-names="dependentNamesFor(item.raw)"
            :orphan-names="orphanNamesFor(item.raw)"
            :update-check="updateChecks.get(item.raw.id)"
            @reload="$emit('reload')"
            @pinned="onPinned"
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

