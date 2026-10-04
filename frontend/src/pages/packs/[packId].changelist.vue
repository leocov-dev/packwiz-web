<route lang="yaml">
meta:
  layout: app
</route>

<script setup lang="ts">
import {useRoute} from "vue-router";
import {buildDataLoader} from "@/composables/data-loader.ts";
import {fetchPackChangelist} from "@/services/changelist.service.ts";
import PackChangelist from "@/components/changelist/PackChangelist.vue";

const route = useRoute<'/packs/[packId].changelist'>()
const packId = computed(() => Number(route.params.packId))

const {isLoading, data: changelist, error} = buildDataLoader(() => fetchPackChangelist(packId.value))
</script>

<template>
  <v-card class="ma-6">
    <v-card-title class="d-flex align-center">
      <v-btn
        icon="mdi-arrow-left"
        variant="text"
        aria-label="Back to pack"
        :to="`/packs/${packId}`"
      />
      <h1 class="ms-2">
        Changelist
      </h1>
    </v-card-title>
    <v-card-text>
      <p class="text-medium-emphasis mb-4">
        What players received over time: the last 10 days by day, recent months by month,
        older changes by year. Changes made while the pack was a draft appear when it was published.
      </p>
      <v-skeleton-loader
        v-if="isLoading && !changelist"
        type="paragraph@3"
      />
      <v-alert
        v-else-if="error || !changelist"
        type="error"
        icon="mdi-alert"
        text="Failed to load the changelist."
      />
      <PackChangelist
        v-else
        :entries="changelist.entries"
      />
    </v-card-text>
  </v-card>
</template>
