<script setup lang="ts">
import type {Mod} from "@/interfaces/pack.ts"
import type {DependentsMap} from "@/lib/mod-filters.ts"
import {planRemoval, removeModsMessage} from "@/lib/mod-selection.ts"
import {removeMods} from "@/services/mods.service.ts"
import {apiErrorMessage} from "@/services/utils.ts"
import {useSnackbarStore} from "@/stores/snackbar.ts"

const {packId, mods, selected, dependents} = defineProps<{
  packId: number
  mods: Mod[]
  selected: ReadonlySet<number>
  dependents: DependentsMap
}>()

const emit = defineEmits<{
  removed: []
}>()

const open = defineModel<boolean>({required: true})

const snackbar = useSnackbarStore()
const includeUnused = ref(true)
const removing = ref(false)
const error = ref("")

const plan = computed(() => planRemoval(mods, selected, dependents, includeUnused.value))
const message = computed(() => removeModsMessage(selected.size, plan.value.unused.length, includeUnused.value))
const unusedLabel = computed(() => `Also remove dependencies no longer needed: ${plan.value.unused.map(m => m.name).join(", ")}`)

watch(open, (value) => {
  if (value) {
    includeUnused.value = true
    error.value = ""
  }
})

const confirm = async () => {
  error.value = ""
  removing.value = true
  try {
    const {removed} = await removeMods(packId, plan.value.ids)
    snackbar.showSnackbar(`Removed ${removed} ${removed === 1 ? "mod" : "mods"}`, "success")
    open.value = false
    emit("removed")
  } catch (e) {
    error.value = apiErrorMessage(e, "Failed to remove mods")
  } finally {
    removing.value = false
  }
}
</script>

<template>
  <v-dialog
    v-model="open"
    persistent
    max-width="500"
  >
    <v-card class="pa-3">
      <v-card-title>Remove Mods</v-card-title>
      <v-card-text>
        <p class="mb-3">
          {{ message }}
        </p>
        <v-checkbox
          v-if="plan.unused.length > 0"
          v-model="includeUnused"
          :label="unusedLabel"
          hide-details
          class="mb-3"
        />
        <v-alert
          v-if="plan.blocked.length > 0"
          type="warning"
          variant="tonal"
          density="compact"
          class="mb-3"
          :text="`Still required, select the mods that need them too: ${plan.blocked.join('; ')}`"
        />
        <v-alert
          v-if="error"
          type="error"
          variant="tonal"
          density="compact"
          :text="error"
        />
      </v-card-text>
      <v-card-actions>
        <v-spacer />
        <v-btn
          text="Cancel"
          variant="text"
          :disabled="removing"
          @click="open = false"
        />
        <v-btn
          :text="`Remove ${plan.ids.length}`"
          color="error"
          :loading="removing"
          :disabled="plan.ids.length === 0 || plan.blocked.length > 0"
          @click="confirm"
        />
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>
