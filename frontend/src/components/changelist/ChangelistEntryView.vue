<script setup lang="ts">
import type {ChangelistEntry, ChangelistModChange} from "@/interfaces/changelist.ts";
import {entryTitle, modFieldLabel, packChangeLine} from "@/lib/changelist.ts";

const {entry, showTitle = true} = defineProps<{
  entry: ChangelistEntry
  showTitle?: boolean
}>()

const COLLAPSED = 8

const packLines = computed(() => entry.pack.map(c => packChangeLine(c, entry.initial)))

const groups = computed(() => [
  {
    key: "added", title: "Added", icon: "mdi-plus-circle-outline", color: "success",
    items: entry.added.map(m => ({slug: m.slug, name: m.name, detail: m.version})),
  },
  {
    key: "changed", title: "Updated", icon: "mdi-update", color: "info",
    items: entry.changed.map(m => ({slug: m.slug, name: m.name, detail: changeDetail(m)})),
  },
  {
    key: "removed", title: "Removed", icon: "mdi-minus-circle-outline", color: "error",
    items: entry.removed.map(m => ({slug: m.slug, name: m.name, detail: m.version})),
  },
].filter(g => g.items.length > 0))

function changeDetail(m: ChangelistModChange): string {
  const parts: string[] = []
  if (m.updated) {
    parts.push(m.fromVersion && m.toVersion && m.fromVersion !== m.toVersion
      ? `${m.fromVersion} → ${m.toVersion}`
      : "new file")
  }
  parts.push(...m.fields.map(modFieldLabel))
  return parts.join(", ")
}

const expanded = ref<Record<string, boolean>>({})
function visible<T>(key: string, items: T[]): T[] {
  return expanded.value[key] ? items : items.slice(0, COLLAPSED)
}
</script>

<template>
  <div>
    <div
      v-if="showTitle"
      class="d-flex flex-wrap align-center ga-2 mb-1"
    >
      <span class="text-subtitle-1 font-weight-medium">{{ entryTitle(entry) }}</span>
      <v-chip
        v-if="entry.initial"
        size="x-small"
        label
        variant="tonal"
      >
        history starts
      </v-chip>
      <v-spacer />
      <span class="text-caption text-medium-emphasis">{{ entry.modCount }} mods</span>
    </div>

    <div
      v-for="line in packLines"
      :key="line.label"
      class="d-flex align-center ga-2 text-body-2"
      :class="{'text-warning': line.target && !entry.initial}"
    >
      <v-icon
        size="small"
        :icon="line.target ? 'mdi-minecraft' : 'mdi-information-outline'"
      />
      <span class="font-weight-medium">{{ line.label }}</span>
      <span v-if="line.detail">{{ line.detail }}</span>
    </div>

    <div
      v-for="group in groups"
      :key="group.key"
      class="mt-2"
    >
      <div class="d-flex align-center ga-1 text-body-2 font-weight-medium">
        <v-icon
          size="small"
          :icon="group.icon"
          :color="group.color"
        />
        {{ group.title }} ({{ group.items.length }})
      </div>
      <ul class="ms-7 text-body-2">
        <li
          v-for="item in visible(group.key, group.items)"
          :key="item.slug"
        >
          {{ item.name }}
          <span
            v-if="item.detail"
            class="text-medium-emphasis"
          >{{ item.detail }}</span>
        </li>
      </ul>
      <v-btn
        v-if="group.items.length > COLLAPSED"
        variant="text"
        size="small"
        density="compact"
        class="ms-5"
        :text="expanded[group.key] ? 'Show less' : `Show all ${group.items.length}`"
        @click="expanded[group.key] = !expanded[group.key]"
      />
    </div>
  </div>
</template>
