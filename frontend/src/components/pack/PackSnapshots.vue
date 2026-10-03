<script setup lang="ts">
import {PackStatus, type PackResponse} from "@/interfaces/pack.ts";
import type {PackSnapshot} from "@/interfaces/snapshot.ts";
import SnapshotDetailDialog from "@/components/pack/SnapshotDetailDialog.vue";
import {buildDataLoader} from "@/composables/data-loader.ts";
import {
  snapshotReasonIcon,
  snapshotReasonLabel,
  snapshotSubject,
  summarizeSnapshot,
} from "@/lib/snapshots.ts";
import {fetchPackSnapshots} from "@/services/snapshots.service.ts";

const {pack} = defineProps<{ pack: PackResponse }>()

const emit = defineEmits<{ reload: [] }>()

const router = useRouter()

const page = ref(1)
const itemsPerPage = ref(25)
const showAbandoned = ref(false)

const selected = ref<PackSnapshot | null>(null)
const showDetail = ref(false)

const headers = [
  {title: '#', key: 'seq', sortable: false},
  {title: 'When', key: 'createdAt', sortable: false},
  {title: 'By', key: 'createdByUsername', sortable: false},
  {title: 'Change', key: 'reason', sortable: false},
  {title: 'Details', key: 'summary', sortable: false},
  {title: '', key: 'status', sortable: false},
]

const {
  isLoading,
  data,
  reload,
} = buildDataLoader(async () => {
  return fetchPackSnapshots(pack.id, page.value, itemsPerPage.value, showAbandoned.value)
})

const snapshots = computed(() => data.value?.snapshots ?? [])
const total = computed(() => data.value?.total ?? 0)

const onUpdateOptions = (options: { page: number, itemsPerPage: number }) => {
  page.value = options.page
  itemsPerPage.value = options.itemsPerPage
  reload()
}

watch(showAbandoned, () => {
  page.value = 1
  reload()
})

const onRowClick = (_event: Event, row: { item: PackSnapshot }) => {
  selected.value = row.item
  showDetail.value = true
}

const rowProps = ({item}: { item: PackSnapshot }) => ({
  class: item.isAbandoned ? 'text-disabled' : '',
})

// a revert changes the pack itself, so the page reloads the pack and this list
const onReverted = () => {
  emit('reload')
}

const backToPack = async () => {
  await router.push({path: `/packs/${pack.id}`})
}
</script>

<template>
  <div class="ma-6">
    <SnapshotDetailDialog
      v-model="showDetail"
      :pack="pack"
      :snapshot="selected"
      @reverted="onReverted"
    />

    <v-card>
      <v-card-title class="d-flex align-baseline">
        <h1 class="me-5">
          {{ pack.name || pack.slug }}
        </h1>
        <h2>History</h2>
        <v-spacer />
        <v-btn
          text="Back to pack"
          prepend-icon="mdi-arrow-left"
          variant="text"
          @click="backToPack"
        />
      </v-card-title>

      <v-card-text>
        <v-alert
          v-if="pack.status === PackStatus.DRAFT"
          class="mb-4"
          type="info"
          variant="tonal"
          text="Draft packs don't record history. Changes made now are captured when the pack is published."
        />

        <v-data-table-server
          :headers="headers"
          :items="snapshots"
          :items-length="total"
          :items-per-page="itemsPerPage"
          :page="page"
          :loading="isLoading"
          :row-props="rowProps"
          hover
          @click:row="onRowClick"
          @update:options="onUpdateOptions"
        >
          <template #top>
            <!-- a plain flex row, not v-toolbar: the toolbar clips the switch thumb on the left -->
            <div class="d-flex align-center px-4 py-2">
              <v-switch
                v-model="showAbandoned"
                label="Show abandoned"
                density="comfortable"
                hide-details
              />
              <v-spacer />
              <v-btn
                icon="mdi-refresh"
                variant="text"
                density="comfortable"
                aria-label="Refresh"
                @click="reload()"
              />
            </div>
          </template>

          <template #[`item.createdAt`]="{ item }">
            {{ new Date(item.createdAt).toLocaleString() }}
          </template>

          <template #[`item.createdByUsername`]="{ item }">
            {{ item.createdByUsername || 'unknown' }}
          </template>

          <template #[`item.reason`]="{ item }">
            <v-chip
              size="small"
              label
              variant="tonal"
              :prepend-icon="snapshotReasonIcon(item.reason)"
              :text="snapshotReasonLabel(item.reason)"
            />
            <span
              v-if="snapshotSubject(item.detail)"
              class="ms-2"
            >
              {{ snapshotSubject(item.detail) }}
            </span>
          </template>

          <template #[`item.summary`]="{ item }">
            {{ summarizeSnapshot(item.summary) }}
          </template>

          <template #[`item.status`]="{ item }">
            <v-chip
              v-if="item.isHead"
              size="small"
              label
              variant="tonal"
              color="primary"
              text="Current"
            />
            <v-chip
              v-if="item.isAbandoned"
              size="small"
              label
              variant="tonal"
              text="Abandoned"
            />
          </template>

          <template #loading>
            <v-skeleton-loader type="table-row@5" />
          </template>

          <template #no-data>
            <div class="d-flex justify-center ma-10">
              No snapshots yet. History starts when the pack is published.
            </div>
          </template>
        </v-data-table-server>
      </v-card-text>
    </v-card>
  </div>
</template>
