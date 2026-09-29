<script setup lang="ts">
import type {Mod} from "@/interfaces/pack.ts"
import {updateModFromSource} from "@/services/mods.service.ts"
import {apiErrorMessage} from "@/services/utils.ts"
import {modSourceLabel} from "@/lib/mod-filters.ts"
import {snapshotMod, type ModSnapshot} from "@/lib/mod-edit.ts"

const {packId, mod, hasUnsavedChanges, pinned, disabled = false} = defineProps<{
  packId: number
  mod: Mod
  hasUnsavedChanges: boolean
  // Saved (baseline) pinned state, not the possibly stale mod prop.
  pinned: boolean
  disabled?: boolean
}>()

// Emits the pre-update snapshot so the parent can compare after its reload.
const emit = defineEmits<{ updated: [before: ModSnapshot], busy: [busy: boolean] }>()

const displayName = computed(() => mod.name || mod.slug)

const confirmOpen = ref(false)
const loading = ref(false)
const errorMsg = ref("")

const sourceLabel = computed(() => modSourceLabel(mod.source))
const confirmText = computed(() => {
  const base = `Update ${displayName.value} from ${sourceLabel.value}? This may change the installed file.`
  return hasUnsavedChanges ? `${base}\nUnsaved edits in the form above will be discarded.` : base
})

const doUpdate = async () => {
  errorMsg.value = ""
  loading.value = true
  emit("busy", true)
  try {
    const before = snapshotMod(mod)
    await updateModFromSource(packId, mod.id)
    emit("updated", before)
  } catch (e) {
    errorMsg.value = apiErrorMessage(e, "Failed to update mod from source")
  } finally {
    loading.value = false
    emit("busy", false)
  }
}
</script>

<template>
  <v-card
    v-if="sourceLabel"
    class="mt-6"
  >
    <v-card-title>
      <h2 class="text-h6">
        Source
      </h2>
    </v-card-title>
    <v-card-text>
      <v-alert
        v-if="errorMsg"
        class="mb-4"
        :text="errorMsg"
        type="error"
        icon="mdi-alert"
        closable
        @click:close="errorMsg = ''"
      />
      <div>Source: {{ sourceLabel }}</div>
      <div v-if="mod.fileName">
        File: {{ mod.fileName }}
      </div>
      <div
        v-if="pinned"
        class="text-medium-emphasis mt-2"
      >
        This mod is pinned. Unpin it (and save) before updating from source.
      </div>
    </v-card-text>
    <v-card-actions>
      <v-btn
        text="Update from source"
        variant="outlined"
        :disabled="loading || disabled || pinned"
        :loading="loading"
        @click="confirmOpen = true"
      />
      <span class="text-medium-emphasis text-body-2 ms-3">Fetches the latest compatible version</span>
    </v-card-actions>

    <ConfirmationDialog
      v-model="confirmOpen"
      title="Update from source?"
      :text="confirmText"
      accept-text="Update"
      @accepted="doUpdate"
    />
  </v-card>
</template>
