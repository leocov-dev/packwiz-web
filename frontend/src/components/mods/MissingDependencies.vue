<script setup lang="ts">
import type {ModDependency} from "@/interfaces/pack.ts";

const {missing} = defineProps<{missing: ModDependency[]}>()


</script>

<template>
  <v-card variant="tonal">
    <v-card-title
      class="mt-2 text-subtitle-2"
    >
      These additional dependencies will be installed
    </v-card-title>

    <v-card-text>
      <v-list>
        <v-list-item
          v-for="mod in missing"
          :key="mod.slug"
        >
          - {{ mod.name }} ({{ mod.fileName }})
        </v-list-item>
      </v-list>
    </v-card-text>
  </v-card>
</template>
