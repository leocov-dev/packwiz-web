<script setup lang="ts">

import type {Mod, UpdateCheckItem} from "@/interfaces/pack.ts";
import {modUpdateBadge} from "@/lib/update-checks.ts";
import {pinMod, removeMod, unpinMod} from "@/services/mods.service.ts";
import {dependencyTooltip, displayVersion, modSideLabel, modVersion, modSourceLabel, removeModMessage} from "@/lib/mod-filters.ts";
import {useSnackbarStore} from "@/stores/snackbar.ts";
import ConfirmationDialog from "@/components/ConfirmationDialog.vue";
import ModEditDialog from "@/components/mods/ModEditDialog.vue";
import axios from "axios";
import {Perm, hasPermission} from "@/lib/permissions.ts";

const {packId, mod, permissions, dependentNames, orphanNames, updateCheck = undefined} = defineProps<{
  packId: number,
  mod: Mod,
  permissions: string[],
  dependentNames: string[],
  orphanNames: string[],
  updateCheck?: UpdateCheckItem,
}>()

const canConfigure = computed(() => hasPermission(permissions, Perm.PackModConfigure))
const canRemove = computed(() => hasPermission(permissions, Perm.PackModRemove))

const emit = defineEmits<{
  reload: []
  pinned: [id: number, value: boolean]
}>()

const snackbar = useSnackbarStore()

// Local optimistic copy: props are read-only, so re-sync when the parent refreshes
// (but never while a request is in flight).
const pinned = ref(mod.pinned)
const pinning = ref(false)
watch(() => mod.pinned, (value) => {
  if (!pinning.value) pinned.value = value
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
const removeText = computed(() => removeModMessage(mod.name, orphanNames, mod.isDependency))
const dependencyHint = computed(() => dependencyTooltip(dependentNames))
// A dependency can only be removed once no mod requires it anymore.
const removeBlocked = computed(() => mod.isDependency && dependentNames.length > 0)
const editHint = computed(() => mod.isDependency
  ? (dependentNames.length > 0 ? dependencyHint.value : `Dependencies can't be edited. ${dependencyHint.value}`)
  : "")
const removeHint = computed(() => removeBlocked.value ? dependencyHint.value : "")
const showEditDialog = ref(false)
const pinTooltip = computed(() => pinned.value ? "Unpin" : "Pin (skip on Update All)")
// uses the optimistic pin state so the badge follows a pin toggle immediately
const updateBadge = computed(() => modUpdateBadge({pinned: pinned.value, version: mod.version}, updateCheck))
const versionText = computed(() => displayVersion(mod))
const hasVersion = computed(() => modVersion(mod) !== "")
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
    emit('pinned', mod.id, !previous)
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
      :title="mod.isDependency ? 'Remove Dependency' : 'Remove Mod'"
      :text="removeText"
      accept-text="Remove"
      @accepted="onRemove"
    />

    <ModEditDialog
      v-model="showEditDialog"
      :pack-id="packId"
      :mod="mod"
      :pinned="pinned"
      @reload="emit('reload')"
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
        tabindex="0"
        :aria-label="`Type: ${typeLabel}`"
        role="img"
        :icon="modTypeIconMap[mod.type] || 'mdi-puzzle-outline'"
      />

      <div
        v-tooltip="mod.fileName"
        tabindex="0"
        class="text-body-1 font-weight-medium"
      >
        {{ mod.name }}
      </div>

      <div
        v-if="versionText"
        v-tooltip="mod.fileName || versionText"
        tabindex="0"
        class="text-body-2 text-medium-emphasis text-truncate mod-version"
      >
        <span class="mod-version-prefix">{{ hasVersion ? "Version " : "File " }}</span>{{ versionText }}
      </div>

      <v-chip
        v-if="sourceLabel"
        size="small"
        label
        variant="tonal"
        :text="sourceLabel"
      />
      <v-chip
        v-if="sideLabel"
        size="small"
        label
        variant="tonal"
        :text="sideLabel"
      />
      <v-chip
        v-if="mod.option?.optional"
        v-tooltip="mod.option?.description || 'Optional mod'"
        tabindex="0"
        size="small"
        label
        color="info"
        variant="tonal"
        text="Optional"
      />
      <v-chip
        v-if="mod.isDependency"
        v-tooltip="dependencyHint"
        tabindex="0"
        size="small"
        label
        color="primary"
        variant="tonal"
        prepend-icon="mdi-graph"
        text="Dependency"
      />

      <v-chip
        v-if="updateBadge?.kind === 'available'"
        v-tooltip="updateBadge.tooltip"
        tabindex="0"
        size="small"
        label
        color="success"
        variant="tonal"
        prepend-icon="mdi-arrow-up-bold-circle-outline"
        text="Update available"
      />
      <v-chip
        v-else-if="updateBadge?.kind === 'pinned'"
        v-tooltip="updateBadge.tooltip"
        tabindex="0"
        size="small"
        label
        variant="text"
        class="text-medium-emphasis"
        prepend-icon="mdi-arrow-up-bold-circle-outline"
        text="Update available (pinned)"
      />
      <v-icon
        v-else-if="updateBadge?.kind === 'error'"
        v-tooltip="updateBadge.tooltip"
        tabindex="0"
        role="img"
        aria-label="Update check failed"
        size="small"
        color="warning"
        icon="mdi-alert-circle-outline"
      />

      <v-spacer />

      <div class="d-flex align-center ga-2">
        <template v-if="!mod.isDependency || mod.pinned">
          <v-btn
            v-if="canConfigure && !mod.isDependency"
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
            tabindex="0"
            icon="mdi-pin"
          />
        </template>

        <template v-if="canConfigure || canRemove">
          <v-tooltip
            v-if="canConfigure"
            :disabled="!mod.isDependency"
            :text="editHint"
            location="top"
          >
            <template #activator="{props: tipProps}">
              <span
                v-bind="tipProps"
                :role="mod.isDependency ? 'group' : undefined"
                :tabindex="mod.isDependency ? 0 : undefined"
                :aria-label="mod.isDependency ? `Edit unavailable. ${editHint}` : undefined"
              >
                <v-btn
                  density="comfortable"
                  variant="tonal"
                  text="Edit"
                  :disabled="mod.isDependency"
                  @click="showEditDialog = true"
                />
              </span>
            </template>
          </v-tooltip>

          <v-tooltip
            v-if="canRemove"
            :disabled="!removeBlocked"
            :text="removeHint"
            location="top"
          >
            <template #activator="{props: tipProps}">
              <span
                v-bind="tipProps"
                :role="removeBlocked ? 'group' : undefined"
                :tabindex="removeBlocked ? 0 : undefined"
                :aria-label="removeBlocked ? `Remove unavailable. ${removeHint}` : undefined"
              >
                <v-btn
                  density="comfortable"
                  color="error"
                  variant="text"
                  icon="mdi-delete-outline"
                  aria-label="Remove mod"
                  :disabled="loading || removeBlocked"
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

<style scoped>
/* No Vuetify utility caps width by ch, or hides text visually (see STYLE_GUIDE.md) */
.mod-version {
  max-width: min(24ch, 100%);
}
.mod-version-prefix {
  position: absolute;
  width: 1px;
  height: 1px;
  overflow: hidden;
  clip-path: inset(50%);
  white-space: nowrap;
}
</style>
