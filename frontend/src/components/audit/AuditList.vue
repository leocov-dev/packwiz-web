<script setup lang="ts">
import {fetchAuditsPaginated} from "@/services/audit.service.ts";
import {buildDataLoader} from "@/composables/data-loader.ts";

const actionSearch = ref('')
const userIdSearch = ref('')
const startDate = ref('')
const endDate = ref('')

const page = ref(1)
const itemsPerPage = ref(20)

const headers = [
  {title: 'Created', key: 'createdAt'},
  {title: 'User', key: 'userId'},
  {title: 'Action', key: 'action'},
  {title: 'Params', key: 'actionParams'},
  {title: 'IP Address', key: 'ipAddress'},
]

const {
  isLoading,
  data,
  reload,
} = buildDataLoader(async () => {
  const userId = userIdSearch.value ? Number(userIdSearch.value) : undefined
  return fetchAuditsPaginated(page.value, itemsPerPage.value, actionSearch.value, userId, startDate.value, endDate.value)
})

const audits = computed(() => data.value?.results || [])
const total = computed(() => data.value?.pagination.total || 0)

const onUpdateOptions = (options: { page: number, itemsPerPage: number }) => {
  page.value = options.page
  itemsPerPage.value = options.itemsPerPage
  reload()
}

watch([actionSearch, userIdSearch, startDate, endDate], () => {
  page.value = 1
  reload()
})

function formatParams(raw: string): string {
  try {
    return JSON.stringify(JSON.parse(raw), null, 2)
  } catch {
    return raw
  }
}
</script>

<template>
  <v-data-table-server
    :headers="headers"
    :items="audits"
    :items-length="total"
    :items-per-page="itemsPerPage"
    :page="page"
    :loading="isLoading"
    @update:options="onUpdateOptions"
  >
    <template #top>
      <v-toolbar
        class="ps-5 pe-5 pt-2 pb-2 d-flex flex-wrap"
        elevation="4"
      >
        <v-text-field
          v-model="actionSearch"
          max-width="220"
          class="me-3"
          density="comfortable"
          placeholder="Search by action"
          prepend-inner-icon="mdi-magnify"
          variant="solo"
          clearable
          hide-details
          @keyup.enter="reload"
          @click:clear="() => { actionSearch = ''; reload() }"
        />
        <v-text-field
          v-model="userIdSearch"
          max-width="140"
          class="me-3"
          density="comfortable"
          placeholder="User ID"
          prepend-inner-icon="mdi-account"
          variant="solo"
          type="number"
          clearable
          hide-details
          @keyup.enter="reload"
          @click:clear="() => { userIdSearch = ''; reload() }"
        />
        <v-text-field
          v-model="startDate"
          max-width="170"
          class="me-3"
          density="comfortable"
          label="From"
          variant="solo"
          type="date"
          clearable
          hide-details
        />
        <v-text-field
          v-model="endDate"
          max-width="170"
          class="me-auto"
          density="comfortable"
          label="To"
          variant="solo"
          type="date"
          clearable
          hide-details
        />
        <v-btn
          icon="mdi-refresh"
          @click="reload()"
        />
      </v-toolbar>
    </template>

    <template #item.createdAt="{ item }">
      {{ new Date(item.createdAt).toLocaleString() }}
    </template>

    <template #item.userId="{ item }">
      <router-link :to="`/admin/users/${item.userId}`">
        {{ item.userId }}
      </router-link>
    </template>

    <template #item.actionParams="{ item }">
      <v-tooltip
        :text="formatParams(item.actionParams)"
        location="bottom"
      >
        <template #activator="{ props }">
          <span
            v-bind="props"
            class="text-truncate d-inline-block"
            style="max-width: 240px"
          >
            {{ item.actionParams }}
          </span>
        </template>
      </v-tooltip>
    </template>

    <template #loading>
      <v-skeleton-loader type="table-row@5" />
    </template>

    <template #no-data>
      <div class="d-flex justify-center ma-10">
        No results.
      </div>
    </template>
  </v-data-table-server>
</template>
