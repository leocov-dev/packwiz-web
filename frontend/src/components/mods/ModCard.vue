<script setup lang="ts">

import type {Mod} from "@/interfaces/pack.ts";
import {removeMod} from "@/services/mods.service.ts";
import ConfirmationDialog from "@/components/ConfirmationDialog.vue";
import axios from "axios";

const {packId, mod, canEdit} = defineProps<{ packId: number, mod: Mod, canEdit: boolean }>()

const emit = defineEmits(['reload'])

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

</script>

<template>
  <v-card class="ma-1 ps-5 pe-5 pt-3 pb-3 elevation-4">
    <ConfirmationDialog
      v-model="showRemoveDialog"
      title="Remove Mod"
      :text="`Are you sure you want to remove ${mod.name}?`"
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

    <div class="d-flex align-center">
      <v-icon
        v-tooltip="mod.type"
        class="me-2"
        :icon="modTypeIconMap[mod.type] || 'mdi-puzzle-outline'"
      />

      <div>
        {{ mod.name }}
      </div>

      <v-spacer />

      <div
        class="ms-4 me-8 text-subtitle-2 text-disabled text-truncate"
      >
        {{ mod.fileName }}
      </div>

      <div class="d-flex justify-end">
        <v-icon
          v-if="mod.side === 'client'"
          v-tooltip="'client'"
          class="me-2"
          icon="mdi-account-outline"
        />
        <v-icon
          v-if="mod.side === 'server'"
          v-tooltip="'server'"
          class="me-2"
          icon="mdi-server-outline"
        />
        <v-icon
          v-if="mod.side === 'both'"
          v-tooltip="'server+client'"
          class="me-2"
          icon="mdi-circle-double"
        />
        <v-icon
          v-if="mod.isDependency"
          v-tooltip="'Dependency'"
          class="me-2"
          icon="mdi-graph"
          color="primary"
        />
        <v-icon
          v-if="!mod.isDependency"
          v-tooltip="mod.pinned ? 'pinned' : 'unpinned'"
          class="me-2"
          :icon="mod.pinned ? 'mdi-pin' : 'mdi-pin-off-outline'"
        />

        <v-btn
          v-if="canEdit"
          density="comfortable"
          color="warning"
          variant="outlined"
          text="Edit"
          :to="`${packId}/mod/${mod.id}`"
          :disabled="mod.isDependency"
        />

        <v-btn
          v-if="canEdit"
          class="ms-2"
          density="comfortable"
          color="error"
          variant="outlined"
          icon="mdi-delete-outline"
          :disabled="loading || mod.isDependency"
          @click="showRemoveDialog = true"
        />
      </div>
    </div>
  </v-card>
</template>
