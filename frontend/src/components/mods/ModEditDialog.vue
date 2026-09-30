<script setup lang="ts">
import type {Mod} from "@/interfaces/pack.ts";
import {changeModOption, changeModSide, fetchOneMod} from "@/services/mods.service.ts";
import {apiErrorMessage} from "@/services/utils.ts";
import {useSnackbarStore} from "@/stores/snackbar.ts";
import ModSourceCard from "@/components/mods/ModSourceCard.vue";
import {
  buildOptionRequest,
  describeSaveFailure,
  describeUpdateResult,
  diffModEdit,
  editValuesFromMod,
  MAX_OPTION_DESCRIPTION,
  runSaveSteps,
  type ModEditValues,
  type SaveStep,
} from "@/lib/mod-edit.ts";

const model = defineModel<boolean>({required: true})

const {packId, mod, pinned} = defineProps<{
  packId: number
  mod: Mod
  // Current (possibly optimistic) pin state from the card.
  pinned: boolean
}>()

const emit = defineEmits<{ reload: [] }>()

const snackbar = useSnackbarStore()

const errorMsg = ref("")
const isValid = ref<boolean | null>(null)
const saving = ref(false)
const updating = ref(false)
const busy = computed(() => saving.value || updating.value)

// Baseline the form is compared against; advanced as individual saves succeed.
const initial = ref<ModEditValues>(editValuesFromMod(mod))
const data = ref<ModEditValues>({...initial.value})

const diff = computed(() => diffModEdit(initial.value, data.value))
const isDirty = computed(() => diff.value.any)
const canSave = computed(() => isDirty.value && isValid.value !== false && !busy.value)

// Start from the latest mod values each time the dialog opens.
watch(model, open => {
  if (!open) return
  errorMsg.value = ""
  initial.value = editValuesFromMod(mod)
  data.value = {...initial.value}
})

const rules = {
  description: (v: string) => (v ?? "").length <= MAX_OPTION_DESCRIPTION
    || `Max ${MAX_OPTION_DESCRIPTION} characters`,
}

const buildSteps = (): SaveStep[] => {
  const steps: SaveStep[] = []
  const target = {...data.value}
  if (diff.value.side) {
    steps.push({
      label: "Side change",
      run: async () => {
        await changeModSide(packId, mod.id, {side: target.side})
        initial.value = {...initial.value, side: target.side}
      },
    })
  }
  if (diff.value.option) {
    steps.push({
      label: "Options",
      run: async () => {
        await changeModOption(packId, mod.id, buildOptionRequest(target))
        initial.value = {
          ...initial.value,
          optional: target.optional,
          description: target.description,
          default: target.default,
        }
      },
    })
  }
  return steps
}

const save = async () => {
  if (!canSave.value) return
  errorMsg.value = ""
  saving.value = true
  try {
    const results = await runSaveSteps(buildSteps(), e => apiErrorMessage(e, "request failed"))
    const failed = results.some(r => !r.ok)
    if (failed) {
      errorMsg.value = describeSaveFailure(results)
    } else {
      snackbar.showSnackbar(`Saved ${mod.name || mod.slug}`, "success")
      model.value = false
    }
    // Some steps may have succeeded even on failure.
    emit("reload")
  } finally {
    saving.value = false
  }
}

const onSourceUpdated = async (changed: boolean) => {
  const name = mod.name || mod.slug
  let message = changed ? `Updated ${name}` : `${name} is already up to date`
  try {
    const fresh = await fetchOneMod(packId, mod.id)
    message = describeUpdateResult(name, changed, fresh.fileName)
  } catch {
    // keep the generic message; the list reload below still refreshes the card
  }
  snackbar.showSnackbar(message, "success")
  emit("reload")
  model.value = false
}
</script>

<template>
  <v-dialog
    v-model="model"
    max-width="600"
    :persistent="isDirty || busy"
    scrollable
  >
    <v-card :loading="busy">
      <v-card-title>{{ mod.name || mod.slug }}</v-card-title>
      <v-card-subtitle>Edit mod</v-card-subtitle>

      <v-card-text>
        <v-alert
          v-if="errorMsg"
          class="mb-4"
          :text="errorMsg"
          type="error"
          variant="tonal"
          closable
          @click:close="errorMsg = ''"
        />

        <v-form
          v-model="isValid"
          @submit.prevent="save"
        >
          <v-select
            v-model="data.side"
            :items="[
              {title: 'Client', value: 'client'},
              {title: 'Server', value: 'server'},
              {title: 'Client + Server', value: 'both'},
            ]"
            label="Side"
          />

          <v-switch
            v-model="data.optional"
            label="Optional (not required by every install)"
            color="primary"
            hide-details
          />

          <v-text-field
            v-if="data.optional"
            v-model="data.description"
            label="Option description"
            class="mt-3"
            :rules="[rules.description]"
          />

          <v-switch
            v-if="data.optional"
            v-model="data.default"
            label="Enabled by default"
            color="primary"
            hide-details
          />
        </v-form>

        <ModSourceCard
          :pack-id="packId"
          :mod="mod"
          :pinned="pinned"
          :disabled="saving"
          :has-unsaved-changes="isDirty"
          @busy="updating = $event"
          @updated="onSourceUpdated"
        />
      </v-card-text>

      <v-card-actions>
        <v-spacer />
        <v-btn
          text="Cancel"
          variant="text"
          :disabled="busy"
          @click="model = false"
        />
        <v-btn
          text="Save"
          color="primary"
          :loading="saving"
          :disabled="!canSave"
          @click="save"
        />
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>
