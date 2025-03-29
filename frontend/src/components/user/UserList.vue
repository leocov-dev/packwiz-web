<script lang="ts">
import {type FiltersConfig, type Filters, reduceFilters} from "@/components/FiltersMenu.vue";

export interface UserListModel {
  search: string,
  filters: Filters,
}
</script>

<script setup lang="ts">
const model = defineModel({required: true, type: Object as () => UserListModel})

const filterConfig: FiltersConfig = {
  "active": {
    title: "Show Active"
  },
  "deactivated": {
    title: "Show Deactivated"
  }
}

const isLoading = ref(false)
const data = ref([])
const reload = () => {}

watch(
  model,
  () => {
    reload()
  },
  {deep: true},
)
</script>

<template>
  <v-data-iterator
    :loading="isLoading"
    :items="data"
    items-per-page="10"
  >
    <template #header>
      <v-toolbar class="ps-5 pe-5 pt-2 pb-2">
        <SearchBar
          v-model="model.search"
          max-width="400"
          class="me-auto"
          density="comfortable"
        />

        <v-btn
          text="New User"
          to="/users/new"
          link
          color="primary"
          variant="flat"
          class="me-3"
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
          v-for="(item, i) in items"
          :key="i"
          cols="12"
          md="4"
          sm="12"
        >
          <PackCard :pack="item.raw" />
        </v-col>
      </v-row>
    </template>

    <template #no-data>
      <div class="d-flex justify-center ma-10">
        No results.
      </div>
    </template>
  </v-data-iterator>
</template>
