<script setup lang="ts">
import type {SnapshotDiff} from "@/interfaces/snapshot.ts";
import {fieldLabel, formatFieldValue} from "@/lib/snapshots.ts";

const {diff} = defineProps<{ diff: SnapshotDiff }>()

const isEmpty = computed(() =>
  diff.pack.length === 0 &&
  diff.added.length === 0 &&
  diff.removed.length === 0 &&
  diff.changed.length === 0,
)
</script>

<template>
  <v-alert
    v-if="isEmpty"
    type="info"
    variant="tonal"
    text="No differences."
  />

  <v-list
    v-else
    lines="two"
  >
    <template v-if="diff.pack.length > 0">
      <v-list-subheader>Pack</v-list-subheader>
      <v-list-item
        v-for="change in diff.pack"
        :key="change.field"
        :title="fieldLabel(change.field, 'pack')"
      >
        <template #subtitle>
          <span class="text-wrap text-break">
            {{ formatFieldValue(change.from) }} → {{ formatFieldValue(change.to) }}
          </span>
        </template>
      </v-list-item>
    </template>

    <template v-if="diff.added.length > 0">
      <v-list-subheader>Added mods</v-list-subheader>
      <v-list-item
        v-for="mod in diff.added"
        :key="mod.slug"
        :title="mod.name || mod.slug"
        :subtitle="mod.version || mod.fileName"
        prepend-icon="mdi-plus-circle-outline"
        base-color="success"
      />
    </template>

    <template v-if="diff.removed.length > 0">
      <v-list-subheader>Removed mods</v-list-subheader>
      <v-list-item
        v-for="mod in diff.removed"
        :key="mod.slug"
        :title="mod.name || mod.slug"
        :subtitle="mod.version || mod.fileName"
        prepend-icon="mdi-minus-circle-outline"
        base-color="error"
      />
    </template>

    <template v-if="diff.changed.length > 0">
      <v-list-subheader>Changed mods</v-list-subheader>
      <v-list-item
        v-for="mod in diff.changed"
        :key="mod.slug"
        :title="mod.name || mod.slug"
        prepend-icon="mdi-pencil-outline"
      >
        <template #subtitle>
          <div
            v-for="change in mod.changes"
            :key="change.field"
            class="text-wrap text-break"
          >
            {{ fieldLabel(change.field, 'mod') }}:
            {{ formatFieldValue(change.from) }} → {{ formatFieldValue(change.to) }}
          </div>
        </template>
      </v-list-item>
    </template>
  </v-list>
</template>
