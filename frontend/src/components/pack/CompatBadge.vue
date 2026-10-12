<script setup lang="ts">
import type {CompatStatus} from "@/interfaces/compat.ts"
import {compatDisplay} from "@/lib/compat.ts"

const {status, errors = 0, warnings = 0, clickable = true} = defineProps<{
  status: CompatStatus
  errors?: number
  warnings?: number
  clickable?: boolean
}>()

const emit = defineEmits<{ (e: 'click'): void }>()

const display = computed(() => compatDisplay(status, {errors, warnings}))
</script>

<template>
  <v-tooltip
    :text="display.label"
    location="bottom"
  >
    <template #activator="{ props }">
      <v-chip
        v-bind="props"
        :color="display.color"
        :variant="display.variant"
        :aria-label="display.label"
        :link="clickable"
        size="small"
        label
        @click="clickable && emit('click')"
      >
        <template #prepend>
          <v-progress-circular
            v-if="display.busy"
            indeterminate
            size="14"
            width="2"
            class="me-1"
          />
          <v-icon
            v-else
            :icon="display.icon"
            size="small"
            class="me-1"
          />
        </template>
        {{ display.text }}
      </v-chip>
    </template>
  </v-tooltip>
</template>
