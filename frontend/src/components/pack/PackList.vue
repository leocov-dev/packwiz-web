<script lang="ts">
import {type FiltersConfig, type Filters, reduceFilters} from "@/components/FiltersMenu.vue";

export interface PackListModel {
  search: string,
  filters: Filters,
}
</script>

<script setup lang="ts">
import {useAuthStore} from "@/stores/auth.ts";
import {fetchAllPacks} from "@/services/packs.service.ts";
import {type PackResponse} from "@/interfaces/pack.ts";
import {buildDataLoader} from "@/composables/data-loader.ts";

const model = defineModel({required: true, type: Object as () => PackListModel})

const authStore = useAuthStore()

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
if (authStore.user?.isAdmin) {
  filterConfig["archived"] = {
    title: "Show Archived"
  }
}

watch(
  model,
  () => {
    reload()
  },
  {deep: true},
)

const page = ref(1)

</script>

<template>
  <v-data-iterator
    v-model:page="page"
    :loading="isLoading"
    :items="data"
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
      <v-row class="ma-2">
        <v-col
          v-for="n in 5"
          :key="n"
          cols="12"
          md="4"
          sm="12"
        >
          <v-skeleton-loader type="heading, paragraph, actions" />
        </v-col>
      </v-row>
    </template>

    <template #default="{ items }">
      <v-row class="ma-2">
        <v-col
          v-for="item in items"
          :key="item.raw.id"
          cols="12"
          md="4"
          sm="12"
        >
          <PackCard
            :pack="item.raw"
            elevation="4"
          />
        </v-col>
      </v-row>
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
</template>
