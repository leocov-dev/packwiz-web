<script setup lang="ts">
import {buildDataLoader} from "@/composables/data-loader.ts";
import {fetchPackAccessSeries} from "@/services/access.service.ts";
import AccessLineChart from "@/components/access/AccessLineChart.vue";

const {packId} = defineProps<{ packId: number }>()

const {isLoading, data, error} = buildDataLoader(() => fetchPackAccessSeries(packId, 7))
</script>

<template>
  <AccessLineChart
    v-if="!isLoading && !error && data"
    :series="data"
    :lines="[{kind: 'success', label: 'Requests', color: 'primary'}]"
    compact
    :height="32"
    :width="120"
  />
</template>
