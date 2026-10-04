<script setup lang="ts">
import type {ChangelistEntry} from "@/interfaces/changelist.ts";
import {changelistSections, entryTitle} from "@/lib/changelist.ts";
import ChangelistEntryView from "@/components/changelist/ChangelistEntryView.vue";

const {entries} = defineProps<{ entries: ChangelistEntry[] }>()

const sections = computed(() => changelistSections(entries))
</script>

<template>
  <div>
    <p
      v-if="entries.length === 0"
      class="text-medium-emphasis"
    >
      No changes recorded yet.
    </p>
    <section
      v-for="section in sections"
      :key="section.title"
      class="mb-4"
    >
      <h3 class="text-overline text-medium-emphasis">
        {{ section.title }}
      </h3>
      <template
        v-for="(entry, i) in section.entries"
        :key="entryTitle(entry)"
      >
        <v-divider
          v-if="i > 0"
          class="my-3"
        />
        <ChangelistEntryView :entry="entry" />
      </template>
    </section>
    <p class="text-caption text-medium-emphasis">
      Each entry is the net change over that period. Dates are UTC.
    </p>
  </div>
</template>
