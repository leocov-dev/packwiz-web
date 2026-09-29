<script setup lang="ts">
import {type Pack, type Mod} from "@/interfaces/pack.ts";
import {changeModOption, changeModSide, pinMod, unpinMod} from "@/services/mods.service.ts";
import {apiErrorMessage} from "@/services/utils.ts";
import {useSnackbarStore} from "@/stores/snackbar.ts";
import {
  buildOptionRequest,
  describeSaveFailure,
  diffModEdit,
  editValuesFromMod,
  MAX_OPTION_DESCRIPTION,
  type ModEditValues,
} from "@/lib/mod-edit.ts";

const {pack, mod} = defineProps<{ pack: Pack, mod: Mod }>()

const emit = defineEmits<{ reload: [] }>()

const router = useRouter()
const snackbar = useSnackbarStore()

const errorMsg = ref("")
const isValid = ref<boolean | null>(null)
const loading = ref(false)
const allowLeave = ref(false)

// Baseline the form is compared against; advanced as individual saves succeed.
const initial = ref<ModEditValues>(editValuesFromMod(mod))
const data = ref<ModEditValues>({...initial.value})

const diff = computed(() => diffModEdit(initial.value, data.value))
const isDirty = computed(() => diff.value.any)
const canSave = computed(() => isDirty.value && isValid.value !== false && !loading.value)

const rules = {
  description: (v: string) => (v ?? "").length <= MAX_OPTION_DESCRIPTION
    || `Max ${MAX_OPTION_DESCRIPTION} characters`,
}

interface SaveStep {
  label: string
  run: () => Promise<void>
}

const buildSteps = (): SaveStep[] => {
  const steps: SaveStep[] = []
  const target = {...data.value}
  if (diff.value.side) {
    steps.push({
      label: "Side",
      run: async () => {
        await changeModSide(pack.id, mod.id, {side: target.side})
        initial.value = {...initial.value, side: target.side}
      },
    })
  }
  if (diff.value.pinned) {
    steps.push({
      label: "Pin change",
      run: async () => {
        await (target.pinned ? pinMod(pack.id, mod.id) : unpinMod(pack.id, mod.id))
        initial.value = {...initial.value, pinned: target.pinned}
      },
    })
  }
  if (diff.value.option) {
    steps.push({
      label: "Options",
      run: async () => {
        await changeModOption(pack.id, mod.id, buildOptionRequest(target))
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

const leaveToPack = async () => {
  allowLeave.value = true
  await router.push({path: `/packs/${pack.id}`})
}

const submitForm = async () => {
  if (!canSave.value) return
  errorMsg.value = ""
  loading.value = true

  const steps = buildSteps()
  const saved: string[] = []
  try {
    for (const [i, step] of steps.entries()) {
      try {
        await step.run()
        saved.push(step.label)
      } catch (e) {
        const remaining = steps.slice(i + 1).map(s => s.label)
        errorMsg.value = describeSaveFailure(saved, step.label, apiErrorMessage(e, "request failed"), remaining)
        return
      }
    }
    snackbar.showSnackbar(`Saved ${mod.name || mod.slug}`, "success")
    await leaveToPack()
  } finally {
    loading.value = false
  }
}

const cancelForm = async () => {
  // Route-leave guard asks for confirmation when dirty.
  await router.push({path: `/packs/${pack.id}`})
}

// Confirmation for leaving with unsaved edits, resolved from the route guard.
const leaveDialog = ref(false)
let resolveLeave: ((ok: boolean) => void) | null = null

const settleLeave = (ok: boolean) => {
  resolveLeave?.(ok)
  resolveLeave = null
}

onBeforeRouteLeave(() => {
  if (allowLeave.value || !isDirty.value) return true
  settleLeave(false)
  return new Promise<boolean>(resolve => {
    resolveLeave = resolve
    leaveDialog.value = true
  })
})
</script>

<template>
  <div class="ma-6">
    <v-card>
      <v-card-title class="d-flex align-center">
        <v-btn
          icon="mdi-arrow-left"
          variant="text"
          aria-label="Back to pack"
          class="me-2"
          :disabled="loading"
          @click="cancelForm"
        />
        <h1 class="text-h5">
          {{ pack.name || pack.slug }} / {{ mod.name || mod.slug }}
        </h1>
      </v-card-title>

      <v-card-subtitle>Edit mod</v-card-subtitle>

      <v-alert
        v-if="errorMsg"
        class="mx-6 mt-4"
        :text="errorMsg"
        type="error"
        icon="mdi-alert"
        closable
        @click:close="errorMsg = ''"
      />

      <v-form
        v-model="isValid"
        class="ma-6"
        @submit.prevent="submitForm"
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
          v-model="data.pinned"
          label="Pinned (don't auto-update)"
          color="primary"
          hide-details
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

        <div class="d-flex justify-end mt-6">
          <v-btn
            text="Cancel"
            :disabled="loading"
            class="me-6"
            @click="cancelForm"
          />
          <v-btn
            text="Save"
            color="primary"
            type="submit"
            :disabled="!canSave"
          />
        </div>
      </v-form>

      <v-overlay
        v-model="loading"
        class="align-center justify-center"
        persistent
        contained
      >
        <v-progress-circular
          color="primary"
          size="64"
          indeterminate
        />
      </v-overlay>
    </v-card>

    <ModSourceCard
      :pack-id="pack.id"
      :mod="mod"
      :has-unsaved-changes="isDirty"
      @updated="emit('reload')"
    />

    <ConfirmationDialog
      v-model="leaveDialog"
      title="Discard unsaved changes?"
      text="You have unsaved changes to this mod. Leave without saving?"
      accept-text="Discard"
      cancel-text="Keep editing"
      @result="settleLeave"
    />
  </div>
</template>
