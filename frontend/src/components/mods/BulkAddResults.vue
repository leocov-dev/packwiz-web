<script setup lang="ts">
import type {BulkAddItem, BulkAddRun, Pack} from "@/interfaces/pack.ts"
import BulkAddReview from "@/components/mods/BulkAddReview.vue"
import {filterItems, groupCounts, isReviewable, sourceLabel, statusMeta, type BulkItemGroup} from "@/lib/bulk-add.ts"

const {pack, run, curseforgeAvailable} = defineProps<{
  pack: Pack
  run: BulkAddRun
  curseforgeAvailable: boolean
}>()

const emit = defineEmits<{
  resolved: [item: BulkAddItem]
}>()

const counts = computed(() => groupCounts(run.items))
const group = ref<BulkItemGroup>(counts.value.review > 0 ? "review" : "all")
const shown = computed(() => filterItems(run.items, group.value))
const reviewingId = ref<number | null>(null)

const groups = computed<{ value: BulkItemGroup, text: string }[]>(() => [
  {value: "all", text: `All (${counts.value.all})`},
  {value: "review", text: `Needs review (${counts.value.review})`},
  {value: "added", text: `Added (${counts.value.added})`},
  {value: "skipped", text: `Skipped (${counts.value.skipped})`},
  {value: "failed", text: `Failed (${counts.value.failed})`},
])

const toggleReview = (item: BulkAddItem) => {
  reviewingId.value = reviewingId.value === item.id ? null : item.id
}

const onResolved = (item: BulkAddItem) => {
  reviewingId.value = null
  emit("resolved", item)
}

const lineAnchor = (position: number) => `bulk-line-${position}`

const showLine = (position: number) => {
  group.value = "all"
  void nextTick(() => document.getElementById(lineAnchor(position))?.scrollIntoView({behavior: "smooth", block: "center"}))
}

const matchText = (item: BulkAddItem): string => {
  if (!item.name) return ""
  const source = sourceLabel(item.source)
  return source ? `${item.name} · ${source}` : item.name
}
</script>

<template>
  <div>
    <v-chip-group
      v-model="group"
      mandatory
      selected-class="text-primary"
      class="mb-2"
      aria-label="Filter lines"
    >
      <v-chip
        v-for="g in groups"
        :key="g.value"
        :value="g.value"
        filter
        label
        variant="tonal"
        :text="g.text"
      />
    </v-chip-group>

    <div
      v-if="shown.length === 0"
      class="text-medium-emphasis mb-4"
    >
      Nothing here
    </div>

    <v-list
      v-else
      class="pa-0 bg-transparent"
    >
      <template
        v-for="item in shown"
        :key="item.id"
      >
        <v-list-item
          :id="lineAnchor(item.position)"
          class="mb-2 elevation-2 bg-surface"
          rounded="lg"
        >
          <template #prepend>
            <span class="text-medium-emphasis me-4">{{ item.position }}</span>
          </template>
          <v-list-item-title class="text-break">
            {{ item.line }}
          </v-list-item-title>
          <v-list-item-subtitle
            v-if="matchText(item)"
            class="text-truncate"
          >
            {{ matchText(item) }}
          </v-list-item-subtitle>
          <v-list-item-subtitle
            v-if="item.reason"
            class="text-wrap"
          >
            {{ item.reason }}
            <a
              v-if="item.conflictsWithPosition > 0"
              href="#"
              class="text-primary"
              @click.prevent="showLine(item.conflictsWithPosition)"
            >Show line {{ item.conflictsWithPosition }}</a>
          </v-list-item-subtitle>
          <template #append>
            <div class="d-flex align-center ga-2">
              <v-chip
                size="small"
                label
                variant="tonal"
                :color="statusMeta(item.status).color"
                :prepend-icon="statusMeta(item.status).icon"
                :text="statusMeta(item.status).label"
              />
              <v-btn
                v-if="isReviewable(item.status)"
                :text="reviewingId === item.id ? 'Close' : 'Review'"
                variant="tonal"
                density="comfortable"
                @click="toggleReview(item)"
              />
            </div>
          </template>
        </v-list-item>

        <div
          v-if="reviewingId === item.id"
          class="px-4 pb-4 mb-2"
        >
          <BulkAddReview
            :pack="pack"
            :run-id="run.id"
            :item="item"
            :curseforge-available="curseforgeAvailable"
            @resolved="onResolved"
          />
        </div>
      </template>
    </v-list>
  </div>
</template>
