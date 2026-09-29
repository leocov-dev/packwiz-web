<script setup lang="ts">
import type {Mod} from "@/interfaces/pack.ts"
import {updateModFromSource} from "@/services/mods.service.ts"
import {apiErrorMessage} from "@/services/utils.ts"
import {modSourceLabel} from "@/lib/mod-filters.ts"
import {useSnackbarStore} from "@/stores/snackbar.ts"

const {packId, mod, hasUnsavedChanges} = defineProps<{
  packId: number
  mod: Mod
  hasUnsavedChanges: boolean
}>()

const emit = defineEmits<{ updated: [] }>()

const snackbar = useSnackbarStore()

const confirmOpen = ref(false)
const loading = ref(false)
const errorMsg = ref("")

const sourceLabel = computed(() => modSourceLabel(mod.source))
const confirmText = computed(() => {
  const base = `Update ${mod.name} from ${sourceLabel.value}? This may change the installed file.`
  return hasUnsavedChanges ? `${base}\nUnsaved edits in the form above will be discarded.` : base
})

const doUpdate = async () => {
  errorMsg.value = ""
  loading.value = true
  try {
    await updateModFromSource(packId, mod.id)
    snackbar.showSnackbar(`Updated ${mod.name} from source`, "success")
    emit("updated")
  } catch (e) {
    errorMsg.value = apiErrorMessage(e, "Failed to update mod from source")
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <v-card
    v-if="sourceLabel"
    class="mt-6"
  >
    <v-card-title>Source</v-card-title>
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
        v-if="mod.pinned"
        class="text-medium-emphasis mt-2"
      >
        This mod is pinned. Unpin it (and save) before updating from source.
      </div>
    </v-card-text>
    <v-card-actions>
      <v-btn
        text="Check for update / Update from source"
        variant="outlined"
        :disabled="loading || mod.pinned"
        :loading="loading"
        @click="confirmOpen = true"
      />
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
