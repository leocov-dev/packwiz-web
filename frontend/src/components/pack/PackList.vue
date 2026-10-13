<script lang="ts">
import {type FiltersConfig, type Filters, reduceFilters} from "@/components/FiltersMenu.vue";

export interface PackListModel {
  search: string,
  filters: Filters,
}
</script>

<script setup lang="ts">
import {fetchAllPacks} from "@/services/packs.service.ts";
import {type PackResponse} from "@/interfaces/pack.ts";
import {buildDataLoader} from "@/composables/data-loader.ts";
import {usePermissions} from "@/composables/usePermissions.ts";
import {Perm} from "@/lib/permissions.ts";

const {can} = usePermissions()
const importOpen = ref(false)

const model = defineModel({required: true, type: Object as () => PackListModel})

const {
  isLoading,
  data,
  reload,
} = buildDataLoader<PackResponse[]>(async () => {
  const activeFilters = reduceFilters(model.value.filters)
  const statusList = activeFilters.filter(f => f !== 'archived')
  const archived = activeFilters.includes('archived')

  const response = await fetchAllPacks(statusList, archived, model.value.search)
  return response.packs || []
})

const filterConfig: FiltersConfig = {
  "published": {
    title: "Show Published"
  },
  "draft": {
    title: "Show Drafts"
  }
}
// the backend only returns archived packs the user can view
filterConfig["archived"] = {
  title: "Show Archived"
}

watch(
  model,
  () => {
    reload()
  },
  {deep: true},
)

const page = ref(1)

// published first, then drafts (sort is stable, so backend order holds within a group)
const sectionOf = (pack: PackResponse) => pack.status === 'draft' ? 'draft' : 'published'
const sectionTitles = {published: 'Published', draft: 'Drafts'}
const sortedPacks = computed(() =>
  [...(data.value ?? [])].sort((a, b) =>
    Number(sectionOf(a) === 'draft') - Number(sectionOf(b) === 'draft')))

</script>

<template>
  <div>
    <ImportPackDialog
      v-model="importOpen"
      @imported="reload()"
    />
    <v-data-iterator
      v-model:page="page"
      :loading="isLoading"
      :items="sortedPacks"
      items-per-page="12"
    >
      <template #header>
        <v-toolbar
          class="ps-5 pe-5 pt-2 pb-2"
          elevation="4"
        >
          <SearchBar
            v-model="model.search"
            max-width="400"
            class="me-auto"
            density="comfortable"
            elevation="4"
          />

          <v-btn
            v-if="can(Perm.PackCreate)"
            text="Import"
            color="primary"
            variant="flat"
            class="me-3"
            elevation="4"
            @click="importOpen = true"
          />

          <v-btn
            text="New Pack"
            to="/packs/new"
            link
            color="primary"
            variant="flat"
            class="me-3"
            elevation="4"
          />

          <FiltersMenu
            v-model="model.filters"
            :config="filterConfig"
          />
          <v-btn
            icon="mdi-refresh"
            @click="reload()"
          />
        </v-toolbar>
      </template>

      <template #loader>
        <div class="pack-grid">
          <v-skeleton-loader
            v-for="n in 6"
            :key="n"
            type="heading, paragraph, actions"
          />
        </div>
      </template>

      <template #default="{ items }">
        <div class="pack-grid">
          <template
            v-for="(item, i) in items"
            :key="item.raw.id"
          >
            <div
              v-if="i === 0 || sectionOf(items[i - 1].raw) !== sectionOf(item.raw)"
              class="pack-grid__section"
            >
              <div class="text-overline text-medium-emphasis">
                {{ sectionTitles[sectionOf(item.raw)] }}
              </div>
              <v-divider />
            </div>
            <PackCard
              :pack="item.raw"
              elevation="4"
            />
          </template>
        </div>
      </template>

      <template #no-data>
        <div class="d-flex justify-center ma-10">
          No results.
        </div>
      </template>

      <template #footer="{ pageCount }">
        <div
          v-if="pageCount > 1"
          class="d-flex justify-center pa-4"
        >
          <v-pagination
            v-model="page"
            :length="pageCount"
          />
        </div>
      </template>
    </v-data-iterator>
  </div>
</template>

<style scoped>
/* fixed card width range; column count follows available width */
.pack-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(360px, 420px));
  justify-content: start;
  gap: 16px;
  padding: 12px;
}

.pack-grid__section {
  grid-column: 1 / -1;
}

@media (max-width: 400px) {
  .pack-grid {
    grid-template-columns: 1fr;
  }
}
</style>
