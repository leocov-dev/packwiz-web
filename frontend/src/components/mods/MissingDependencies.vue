<script setup lang="ts">
import type {ModDependency} from "@/interfaces/pack.ts";

const {missing} = defineProps<{missing: ModDependency[]}>()


</script>

<template>
  <v-card variant="tonal">
    <v-card-title class="mt-2 text-subtitle-2">
      These additional dependencies will be installed
    </v-card-title>

    <v-card-text class="d-flex flex-wrap ga-2">
      <v-chip
        v-for="mod in missing"
        :key="mod.slug"
        v-tooltip="mod.fileName"
        size="small"
        label
        color="primary"
        variant="tonal"
        prepend-icon="mdi-graph"
        :text="mod.name"
      />
    </v-card-text>
  </v-card>
</template>
