<script setup lang="ts">

import type {Mod} from "@/interfaces/pack.ts";
import {pinMod, removeMod, unpinMod} from "@/services/mods.service.ts";
import {dependencyTooltip, modSideLabel, modSourceLabel, removeModMessage} from "@/lib/mod-filters.ts";
import {useSnackbarStore} from "@/stores/snackbar.ts";
import ConfirmationDialog from "@/components/ConfirmationDialog.vue";
import axios from "axios";

const {packId, mod, canEdit, dependentNames, orphanNames} = defineProps<{
  packId: number,
  mod: Mod,
  canEdit: boolean,
  dependentNames: string[],
  orphanNames: string[],
}>()

const emit = defineEmits(['reload'])

const snackbar = useSnackbarStore()

// Local optimistic copy: props are read-only, so re-sync when the parent refreshes.
const pinned = ref(mod.pinned)
const pinning = ref(false)
watch(() => mod.pinned, (value) => {
  pinned.value = value
})

const showRemoveDialog = ref(false)
const loading = ref(false)
const error = ref(false)
const errorMsg = ref("")

const modTypeIconMap: {[key: string]: string} = {
  "mods": "mdi-shield-sword-outline",
  "resourcepacks": "mdi-package-variant-closed",
  "shaderpacks": "mdi-crystal-ball",
  "plugins": "mdi-power-socket-us",
}

const handleError = (e: unknown, fallback: string) => {
  error.value = true
  if (axios.isAxiosError(e)) {
    errorMsg.value = e.response?.data?.error || fallback
  } else {
    errorMsg.value = String(e)
  }
}

const onRemove = async () => {
  loading.value = true
  error.value = false
  try {
    await removeMod(packId, mod.id)
    emit('reload')
  } catch (e) {
    handleError(e, "Failed to remove mod")
  } finally {
    loading.value = false
  }
}

const sourceLabel = computed(() => modSourceLabel(mod.source))
const sideLabel = computed(() => modSideLabel(mod.side))
const removeText = computed(() => removeModMessage(mod.name, orphanNames))
const dependencyHint = computed(() => dependencyTooltip(dependentNames))
const modRoute = computed(() => `/packs/${packId}/mod/${mod.id}`)
const pinTooltip = computed(() => pinned.value ? "Unpin" : "Pin (skip on Update All)")
const typeLabel = computed(() => mod.type || "mod")

const onTogglePin = async () => {
  const previous = pinned.value
  pinned.value = !previous
  pinning.value = true
  try {
    if (previous) {
      await unpinMod(packId, mod.id)
    } else {
      await pinMod(packId, mod.id)
    }
  } catch (e) {
    pinned.value = previous
    const detail = axios.isAxiosError(e) ? e.response?.data?.error : undefined
    snackbar.showSnackbar(detail || `Failed to ${previous ? "unpin" : "pin"} ${mod.name}`, 'error')
  } finally {
    pinning.value = false
  }
}

</script>

<template>
  <v-card class="ma-1 ps-5 pe-5 pt-3 pb-3 elevation-4">
    <ConfirmationDialog
      v-model="showRemoveDialog"
      title="Remove Mod"
      :text="removeText"
      accept-text="Remove"
      @accepted="onRemove"
    />

    <v-alert
      v-if="error"
      class="mb-3"
      :text="'Error: ' + (errorMsg || 'something went wrong')"
      type="error"
      icon="mdi-alert"
      density="compact"
      closable
      @click:close="error = false"
    />

    <div class="d-flex flex-wrap align-center ga-2">
      <v-icon
        v-tooltip="typeLabel"
        :aria-label="`Type: ${typeLabel}`"
        role="img"
        :icon="modTypeIconMap[mod.type] || 'mdi-puzzle-outline'"
      />

      <div
        v-tooltip="mod.fileName"
        class="text-body-1 font-weight-medium"
      >
        {{ mod.name }}
      </div>

      <v-chip
        v-if="sourceLabel"
        size="small"
        label
        :text="sourceLabel"
      />
      <v-chip
        v-if="sideLabel"
        size="small"
        label
        variant="outlined"
        :text="sideLabel"
      />
      <v-chip
        v-if="mod.option?.optional"
        v-tooltip="mod.option?.description || 'Optional mod'"
        size="small"
        label
        color="info"
        variant="tonal"
        text="Optional"
      />
      <v-chip
        v-if="mod.isDependency"
        v-tooltip="dependencyHint"
        size="small"
        label
        color="primary"
        variant="tonal"
        prepend-icon="mdi-graph"
        text="Dependency"
      />

      <v-spacer />

      <div class="d-flex align-center ga-2">
        <template v-if="!mod.isDependency">
          <v-btn
            v-if="canEdit"
            v-tooltip="pinTooltip"
            density="comfortable"
            variant="text"
            :icon="pinned ? 'mdi-pin' : 'mdi-pin-off-outline'"
            :color="pinned ? 'primary' : undefined"
            :aria-label="pinTooltip"
            :aria-pressed="pinned"
            :disabled="pinning"
            @click="onTogglePin"
          />
          <v-icon
            v-else-if="pinned"
            v-tooltip="'Pinned (skipped on Update All)'"
            aria-label="Pinned"
            role="img"
            icon="mdi-pin"
          />
        </template>

        <template v-if="canEdit">
          <v-tooltip
            :disabled="!mod.isDependency"
            :text="dependencyHint"
            location="top"
          >
            <template #activator="{props: tipProps}">
              <span
                v-bind="tipProps"
                :tabindex="mod.isDependency ? 0 : undefined"
                :aria-label="mod.isDependency ? `Edit unavailable. ${dependencyHint}` : undefined"
              >
                <v-btn
                  density="comfortable"
                  color="warning"
                  variant="outlined"
                  text="Edit"
                  :to="modRoute"
                  :disabled="mod.isDependency"
                />
              </span>
            </template>
          </v-tooltip>

          <v-tooltip
            :disabled="!mod.isDependency"
            :text="dependencyHint"
            location="top"
          >
            <template #activator="{props: tipProps}">
              <span
                v-bind="tipProps"
                :tabindex="mod.isDependency ? 0 : undefined"
                :aria-label="mod.isDependency ? `Remove unavailable. ${dependencyHint}` : undefined"
              >
                <v-btn
                  density="comfortable"
                  color="error"
                  variant="outlined"
                  icon="mdi-delete-outline"
                  aria-label="Remove mod"
                  :disabled="loading || mod.isDependency"
                  @click="showRemoveDialog = true"
                />
              </span>
            </template>
          </v-tooltip>
        </template>
      </div>
    </div>
  </v-card>
</template>
