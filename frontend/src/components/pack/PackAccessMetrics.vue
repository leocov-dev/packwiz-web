<script setup lang="ts">
import type {PackResponse} from "@/interfaces/pack.ts";
import {buildDataLoader} from "@/composables/data-loader.ts";
import {fetchPackAccessRecent, fetchPackAccessSummary} from "@/services/access.service.ts";
import AccessLineChart from "@/components/access/AccessLineChart.vue";
import {seriesTotal} from "@/lib/access.ts";

const {pack} = defineProps<{ pack: PackResponse }>()

const rangeOptions = [
  {title: 'Last 7 days', value: 7},
  {title: 'Last 30 days', value: 30},
  {title: 'Last 90 days', value: 90},
]
const days = ref(7)

const page = ref(1)
const itemsPerPage = ref(25)

const {
  isLoading: summaryLoading,
  data: summary,
  error: summaryError,
  reload: reloadSummary,
} = buildDataLoader(() => fetchPackAccessSummary(pack.id, days.value))

const {
  isLoading: recentLoading,
  data: recent,
  reload: reloadRecent,
} = buildDataLoader(() => fetchPackAccessRecent(pack.id, days.value, page.value, itemsPerPage.value))

const hasData = computed(() => seriesTotal(summary.value?.series ?? [], "success") > 0)

const recentHeaders = [
  {title: 'Time', key: 'createdAt', sortable: false},
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

const formatTime = (iso: string) => new Date(iso).toLocaleString()
</script>

<template>
  <div class="ma-6">
    <v-card>
      <v-card-title class="d-flex flex-wrap align-center">
        <v-btn
          icon="mdi-arrow-left"
          variant="text"
          class="me-2"
          aria-label="Back to pack"
          :to="`/packs/${pack.id}`"
        />
        <h1 class="me-auto">
          {{ pack.name }} · Access
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
        <p class="text-medium-emphasis mb-4">
          Successful requests for this pack's pack.toml.
        </p>

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
              color="primary"
            >
              Requests: {{ summary.total }}
            </v-chip>
            <v-chip
              label
              variant="tonal"
            >
              Unique IPs: {{ summary.uniqueIps }}
            </v-chip>
            <v-chip
              label
              variant="tonal"
            >
              Unique users: {{ summary.uniqueUsers }}
            </v-chip>
          </div>

          <AccessLineChart
            v-if="hasData"
            :series="summary.series"
            :lines="[{kind: 'success', label: 'Requests', color: 'primary'}]"
          />
          <p
            v-else
            class="text-medium-emphasis"
          >
            No requests in this range.
          </p>

          <v-row class="mt-4">
            <v-col
              cols="12"
              md="6"
            >
              <h2 class="text-subtitle-1 mb-2">
                Top users
              </h2>
              <v-table density="compact">
                <tbody>
                  <tr
                    v-for="user in summary.topUsers"
                    :key="user.userId"
                  >
                    <td>{{ user.username }}</td>
                    <td class="text-end">
                      {{ user.count }}
                    </td>
                  </tr>
                  <tr v-if="summary.topUsers.length === 0">
                    <td class="text-medium-emphasis">
                      No authenticated requests.
                    </td>
                  </tr>
                </tbody>
              </v-table>
            </v-col>
            <v-col
              cols="12"
              md="6"
            >
              <h2 class="text-subtitle-1 mb-2">
                Top IP addresses
              </h2>
              <v-table density="compact">
                <tbody>
                  <tr
                    v-for="ip in summary.topIps"
                    :key="ip.ipAddress"
                  >
                    <td>{{ ip.ipAddress }}</td>
                    <td class="text-end">
                      {{ ip.count }}
                    </td>
                  </tr>
                  <tr v-if="summary.topIps.length === 0">
                    <td class="text-medium-emphasis">
                      No requests.
                    </td>
                  </tr>
                </tbody>
              </v-table>
            </v-col>
          </v-row>
        </template>
      </v-card-text>

      <v-divider />

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
        <template #[`item.username`]="{ item }">
          {{ item.username || 'Public' }}
        </template>
      </v-data-table-server>
    </v-card>
  </div>
</template>
