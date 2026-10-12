<script setup lang="ts">
import type {CompatResponse, CompatStatus} from "@/interfaces/compat.ts"
import CompatBadge from "@/components/pack/CompatBadge.vue"
import {
  compatUnscannedText,
  findingTitle,
  formatScannedAgo,
  groupFindings,
  severityDisplay,
} from "@/lib/compat.ts"

const {compat, status, loading, error, canRescan, rescanning} = defineProps<{
  compat: CompatResponse | null
  status: CompatStatus
  loading: boolean
  error: string
  canRescan: boolean
  rescanning: boolean
}>()

const model = defineModel<boolean>({required: true})
const emit = defineEmits<{ (e: 'rescan'): void }>()

const groups = computed(() => groupFindings(compat?.findings ?? []))
const unscannedText = computed(() => compatUnscannedText(compat?.unscanned ?? 0))
const scannedText = computed(() => formatScannedAgo(compat?.scannedAt))
const isEmpty = computed(() => !!compat && groups.value.length === 0 && (status === 'ok' || status === 'stale'))
</script>

<template>
  <v-dialog
    v-model="model"
    max-width="760"
    scrollable
  >
    <v-card class="pa-3">
      <v-card-title class="d-flex align-center ga-3">
        Mod compatibility
        <CompatBadge
          :status="status"
          :errors="compat?.counts.errors"
          :warnings="compat?.counts.warnings"
          :clickable="false"
        />
      </v-card-title>
      <v-card-subtitle>{{ scannedText }}</v-card-subtitle>

      <v-card-text>
        <v-alert
          type="info"
          variant="tonal"
          density="compact"
          class="mb-3"
          text="Informational only. Based on mod metadata, so it does not guarantee the pack launches. Client-only mods are not verified."
        />
        <v-progress-linear
          v-if="loading"
          indeterminate
          class="mb-3"
        />
        <v-alert
          v-if="error"
          type="error"
          variant="tonal"
          density="compact"
          class="mb-3"
          :text="error"
        />
        <v-alert
          v-if="compat?.runError"
          type="error"
          variant="tonal"
          density="compact"
          class="mb-3"
          :text="`Last scan problem: ${compat.runError}`"
        />
        <p
          v-if="unscannedText"
          class="text-medium-emphasis mb-3"
        >
          {{ unscannedText }}
        </p>
        <p
          v-if="status === 'stale'"
          class="text-medium-emphasis mb-3"
        >
          Results are out of date. Rescan to refresh them.
        </p>
        <p
          v-if="compat && status === 'unknown'"
          class="text-medium-emphasis"
        >
          This pack has not been scanned yet.
        </p>
        <p
          v-if="isEmpty"
          class="text-medium-emphasis"
        >
          No problems found
        </p>

        <section
          v-for="group in groups"
          :key="group.severity"
          class="mb-4"
        >
          <h3 class="text-subtitle-1 d-flex align-center ga-2 mb-2">
            <v-icon
              :icon="severityDisplay(group.severity).icon"
              :color="severityDisplay(group.severity).color"
            />
            {{ severityDisplay(group.severity).title }} ({{ group.total }})
          </h3>
          <v-card
            v-for="mod in group.mods"
            :key="mod.key"
            variant="outlined"
            class="mb-2 pa-3"
          >
            <div class="font-weight-medium mb-1">
              {{ mod.slug }}
            </div>
            <div
              v-for="(finding, i) in mod.findings"
              :key="`${finding.code}-${i}`"
              class="mb-2"
            >
              <div class="text-body-2 font-weight-medium">
                {{ findingTitle(finding.code) }}
              </div>
              <div class="text-body-2 text-medium-emphasis text-break">
                {{ finding.message }}
              </div>
              <div class="d-flex flex-wrap ga-1 mt-1">
                <v-chip
                  v-for="(value, key) in finding.detail"
                  :key="key"
                  size="x-small"
                  variant="tonal"
                  label
                >
                  {{ key }}: {{ value }}
                </v-chip>
              </div>
            </div>
          </v-card>
        </section>
      </v-card-text>

      <v-card-actions>
        <v-btn
          v-if="canRescan"
          prepend-icon="mdi-magnify-scan"
          text="Rescan"
          :loading="rescanning || status === 'queued' || status === 'scanning'"
          :disabled="rescanning || status === 'queued' || status === 'scanning'"
          @click="emit('rescan')"
        />
        <v-spacer />
        <v-btn
          text="Close"
          @click="model = false"
        />
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>
