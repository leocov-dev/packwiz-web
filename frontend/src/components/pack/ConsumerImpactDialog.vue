<script setup lang="ts">
import {fetchPackAccessSummary} from "@/services/access.service.ts";
import type {PackAccessSummary} from "@/interfaces/access.ts";

const STATS_DAYS = 30

const model = defineModel<boolean>({required: true})

const {packId, title, acceptText = "Continue", danger = false} = defineProps<{
  packId: number
  title: string
  acceptText?: string
  /** Use the warning colour for the accept button. */
  danger?: boolean
}>()

const emit = defineEmits(["accepted"])

const stats = ref<PackAccessSummary | null>(null)
const statsError = ref(false)

// load fresh numbers every time the dialog opens
watch(model, async (open) => {
  if (!open) return
  stats.value = null
  statsError.value = false
  try {
    stats.value = await fetchPackAccessSummary(packId, STATS_DAYS)
  } catch {
    statsError.value = true
  }
}, {immediate: true})

const accept = () => {
  model.value = false
  emit("accepted")
}
</script>

<template>
  <v-dialog
    v-model="model"
    persistent
    max-width="640"
    scrollable
  >
    <v-card :title="title">
      <v-card-text>
        <slot />

        <v-alert
          class="mt-4"
          variant="tonal"
          density="compact"
          :type="stats && stats.total > 0 ? 'warning' : 'info'"
          icon="mdi-account-group"
        >
          <template v-if="statsError">
            Couldn't load who uses this pack.
          </template>
          <template v-else-if="!stats">
            Loading who uses this pack...
          </template>
          <template v-else-if="stats.total === 0">
            Nobody has synced this pack in the last {{ STATS_DAYS }} days.
          </template>
          <template v-else>
            In the last {{ STATS_DAYS }} days this pack was synced
            <strong>{{ stats.total }}</strong> times by
            <strong>{{ stats.uniqueUsers }}</strong> signed-in users from
            <strong>{{ stats.uniqueIps }}</strong> IP addresses.
          </template>
        </v-alert>
      </v-card-text>

      <v-card-actions>
        <v-spacer />
        <v-btn
          variant="tonal"
          text="Cancel"
          @click="model = false"
        />
        <v-btn
          :color="danger ? 'warning' : 'primary'"
          variant="flat"
          :text="acceptText"
          @click="accept"
        />
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>
