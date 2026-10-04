<script setup lang="ts">
import type {AccessOutcome} from "@/interfaces/access.ts";
import {buildDataLoader} from "@/composables/data-loader.ts";
import {
  type AccessSource,
  fetchSystemAccessRecent,
  fetchSystemAccessSummary,
} from "@/services/access.service.ts";
import AccessLineChart from "@/components/access/AccessLineChart.vue";

const {source = 'pack-access', title = 'Pack Access'} = defineProps<{
  source?: AccessSource
  title?: string
}>()

// pack pages only show pack.toml syncs, so download rows don't link there
const linkToPack = computed(() => source === 'pack-access')

const rangeOptions = [
  {title: 'Last 7 days', value: 7},
  {title: 'Last 30 days', value: 30},
  {title: 'Last 90 days', value: 90},
]
const outcomeOptions: { title: string, value: AccessOutcome }[] = [
  {title: 'All requests', value: 'all'},
  {title: 'Failures only', value: 'failure'},
  {title: 'Successes only', value: 'success'},
]
const days = ref(30)
const outcome = ref<AccessOutcome>('failure')

const page = ref(1)
const itemsPerPage = ref(25)

const {
  isLoading: summaryLoading,
  data: summary,
  error: summaryError,
  reload: reloadSummary,
} = buildDataLoader(() => fetchSystemAccessSummary(days.value, source))

const {
  isLoading: recentLoading,
  data: recent,
  reload: reloadRecent,
} = buildDataLoader(() => fetchSystemAccessRecent(
  days.value, page.value, itemsPerPage.value, outcome.value, undefined, source,
))

const packHeaders = [
  {title: 'Pack', key: 'name'},
  {title: 'Successful', key: 'success', align: 'end' as const},
  {title: 'Failed', key: 'failure', align: 'end' as const},
]
const recentHeaders = [
  {title: 'Time', key: 'createdAt', sortable: false},
  {title: 'Pack', key: 'packSlug', sortable: false},
  {title: 'Result', key: 'statusCode', sortable: false},
  {title: 'User', key: 'username', sortable: false},
  {title: 'IP Address', key: 'ipAddress', sortable: false},
  {title: 'Client', key: 'userAgent', sortable: false},
]
const recentTotal = computed(() => recent.value?.pagination.total ?? 0)

const onUpdateOptions = (options: { page: number, itemsPerPage: number }) => {
  page.value = options.page
  itemsPerPage.value = options.itemsPerPage
  reloadRecent()
}

watch(days, () => {
  page.value = 1
  reloadSummary()
  reloadRecent()
})
watch(outcome, () => {
  page.value = 1
  reloadRecent()
})

const formatTime = (iso: string) => new Date(iso).toLocaleString()
</script>

<template>
  <v-card>
    <v-card-title class="d-flex flex-wrap align-center">
      <h1 class="me-auto">
        {{ title }}
      </h1>
      <v-select
        v-model="days"
        :items="rangeOptions"
        density="comfortable"
        max-width="200"
        hide-details
        aria-label="Time range"
      />
    </v-card-title>

    <v-divider />

    <v-card-text>
      <v-skeleton-loader
        v-if="summaryLoading && !summary"
        type="article"
      />
      <v-alert
        v-else-if="summaryError || !summary"
        type="error"
        icon="mdi-alert"
        text="Failed to load access metrics."
      />
      <template v-else>
        <div class="d-flex flex-wrap ga-3 mb-4">
          <v-chip
            label
            variant="tonal"
            color="success"
          >
            Successful: {{ summary.success }}
          </v-chip>
          <v-chip
            label
            variant="tonal"
            color="error"
          >
            Failed: {{ summary.failure }}
          </v-chip>
          <v-chip
            label
            variant="tonal"
          >
            Unique IPs: {{ summary.uniqueIps }}
          </v-chip>
        </div>

        <AccessLineChart
          :series="summary.series"
          :lines="[
            {kind: 'success', label: 'Successful', color: 'success'},
            {kind: 'failure', label: 'Failed', color: 'error'},
          ]"
        />

        <v-row class="mt-4">
          <v-col
            cols="12"
            md="7"
          >
            <h2 class="text-subtitle-1 mb-2">
              By pack
            </h2>
            <v-data-table
              :headers="packHeaders"
              :items="summary.packs"
              density="compact"
              items-per-page="10"
            >
              <template #[`item.name`]="{ item }">
                <router-link
                  v-if="linkToPack"
                  :to="`/packs/${item.packId}/access`"
                >
                  {{ item.name }}
                </router-link>
                <span v-else>{{ item.name }}</span>
              </template>
            </v-data-table>
          </v-col>
          <v-col
            cols="12"
            md="5"
          >
            <h2 class="text-subtitle-1 mb-2">
              Top failing IP addresses
            </h2>
            <v-table density="compact">
              <tbody>
                <tr
                  v-for="ip in summary.topFailedIps"
                  :key="ip.ipAddress"
                >
                  <td>{{ ip.ipAddress }}</td>
                  <td class="text-end">
                    {{ ip.count }}
                  </td>
                </tr>
                <tr v-if="summary.topFailedIps.length === 0">
                  <td class="text-medium-emphasis">
                    No failed requests.
                  </td>
                </tr>
              </tbody>
            </v-table>
          </v-col>
        </v-row>
      </template>
    </v-card-text>

    <v-divider />

    <v-toolbar
      class="ps-5 pe-5"
      density="comfortable"
    >
      <v-select
        v-model="outcome"
        :items="outcomeOptions"
        density="comfortable"
        max-width="220"
        hide-details
        aria-label="Result filter"
      />
    </v-toolbar>

    <v-data-table-server
      :headers="recentHeaders"
      :items="recent?.results ?? []"
      :items-length="recentTotal"
      :items-per-page="itemsPerPage"
      :page="page"
      :loading="recentLoading"
      @update:options="onUpdateOptions"
    >
      <template #[`item.createdAt`]="{ item }">
        {{ formatTime(item.createdAt) }}
      </template>
      <template #[`item.packSlug`]="{ item }">
        {{ item.packName || item.packSlug }}
      </template>
      <template #[`item.statusCode`]="{ item }">
        <v-chip
          size="small"
          label
          variant="tonal"
          :color="item.success ? 'success' : 'error'"
        >
          {{ item.statusCode }}
        </v-chip>
      </template>
      <template #[`item.username`]="{ item }">
        {{ item.username || '-' }}
      </template>
    </v-data-table-server>
  </v-card>
</template>
