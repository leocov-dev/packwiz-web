<script setup lang="ts">
import type {UpdateAllResponse} from "@/interfaces/pack.ts"
import {skipReasonLabel, summarizeUpdateAll} from "@/lib/update-summary.ts"

const {result} = defineProps<{ result: UpdateAllResponse }>()
const model = defineModel<boolean>({required: true})

const summary = computed(() => summarizeUpdateAll(result))
</script>

<template>
  <v-dialog
    v-model="model"
    max-width="600"
    scrollable
  >
    <v-card class="pa-3">
      <v-card-title>Update All Results</v-card-title>
      <v-card-subtitle>{{ summary }}</v-card-subtitle>
      <v-card-text>
        <template v-if="result.updated.length">
          <v-list density="compact">
            <v-list-subheader>Updated</v-list-subheader>
            <v-list-item
              v-for="item in result.updated"
              :key="item.modId"
              prepend-icon="mdi-check-circle-outline"
              :title="item.name || item.slug"
              :subtitle="item.fileName"
            />
          </v-list>
        </template>
        <template v-if="result.failed.length">
          <v-list density="compact">
            <v-list-subheader>Failed</v-list-subheader>
            <v-list-item
              v-for="item in result.failed"
              :key="item.modId"
              prepend-icon="mdi-alert-circle-outline"
            >
              <v-list-item-title>{{ item.name || item.slug }}</v-list-item-title>
              <div class="text-body-2 text-medium-emphasis text-break">
                {{ item.error }}
              </div>
            </v-list-item>
          </v-list>
        </template>
        <template v-if="result.skipped.length">
          <v-list density="compact">
            <v-list-subheader>Skipped</v-list-subheader>
            <v-list-item
              v-for="item in result.skipped"
              :key="item.modId"
              prepend-icon="mdi-pin-outline"
              :title="item.name || item.slug"
              :subtitle="skipReasonLabel(item.reason)"
            />
          </v-list>
        </template>
      </v-card-text>
      <v-card-actions>
        <v-spacer />
        <v-btn
          color="primary"
          variant="flat"
          @click="model = false"
        >
          Close
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

