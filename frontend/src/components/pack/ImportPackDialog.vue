<script setup lang="ts">
import type {ImportPackResponse} from "@/interfaces/pack.ts";
import {importPack} from "@/services/packs.service.ts";
import {apiErrorMessage} from "@/services/utils.ts";

const open = defineModel<boolean>({required: true})

const emit = defineEmits<{
  imported: [result: ImportPackResponse]
}>()

const url = ref('')
const isValid = ref(false)
const isSubmitting = ref(false)
const error = ref('')
const result = ref<ImportPackResponse | null>(null)

const rules = {
  url: (v: string) => {
    try {
      const parsed = new URL(v.trim())
      return ['http:', 'https:'].includes(parsed.protocol) || 'Must be an http(s) url'
    } catch {
      return 'Must be a valid url'
    }
  },
}

watch(open, (isOpen) => {
  if (!isOpen) {
    return
  }
  url.value = ''
  error.value = ''
  result.value = null
})

const submit = async () => {
  if (!isValid.value || isSubmitting.value) {
    return
  }

  isSubmitting.value = true
  error.value = ''
  try {
    result.value = await importPack({url: url.value.trim()})
    emit('imported', result.value)
  } catch (e) {
    error.value = apiErrorMessage(e, 'Failed to import the pack.')
  } finally {
    isSubmitting.value = false
  }
}
</script>

<template>
  <v-dialog
    v-model="open"
    persistent
    max-width="700"
  >
    <v-card class="pa-3">
      <v-card-title>Import Pack</v-card-title>

      <v-card-text v-if="!result">
        <p class="mb-6">
          Imports an existing packwiz pack from the address of its pack.toml.
          Creates a new draft pack and makes you its owner. Mods are imported;
          other files in the pack, like configs, are not.
        </p>

        <v-form
          v-model="isValid"
          @submit.prevent="submit"
        >
          <v-text-field
            v-model="url"
            label="pack.toml URL"
            placeholder="https://example.com/mypack/pack.toml"
            :rules="[rules.url]"
            :disabled="isSubmitting"
            autofocus
          />
        </v-form>

        <v-alert
          v-if="error"
          class="mt-4"
          type="error"
          variant="tonal"
          :text="error"
        />
      </v-card-text>

      <v-card-text v-else>
        <v-alert
          type="success"
          variant="tonal"
          class="mb-4"
          :text="`Imported ${result.modsImported} mods as ${result.name}`"
        />

        <v-alert
          v-if="result.warnings.length"
          type="warning"
          variant="tonal"
          class="mb-4"
          title="Warnings"
        >
          <div
            v-for="item in result.warnings"
            :key="item"
            class="text-break"
          >
            {{ item }}
          </div>
        </v-alert>

        <v-alert
          v-if="result.skippedMods.length"
          type="info"
          variant="tonal"
          class="mb-4"
          title="Mods not imported"
        >
          <div
            v-for="item in result.skippedMods"
            :key="item"
            class="text-break"
          >
            {{ item }}
          </div>
        </v-alert>

        <v-alert
          v-if="result.skippedFiles.length"
          type="info"
          variant="tonal"
          title="Files not imported"
        >
          <div
            v-for="item in result.skippedFiles"
            :key="item"
            class="text-break"
          >
            {{ item }}
          </div>
        </v-alert>
      </v-card-text>

      <v-card-actions>
        <v-spacer />
        <template v-if="!result">
          <v-btn
            text="Cancel"
            variant="text"
            :disabled="isSubmitting"
            @click="open = false"
          />
          <v-btn
            text="Import"
            color="primary"
            :disabled="!isValid"
            :loading="isSubmitting"
            @click="submit"
          />
        </template>
        <template v-else>
          <v-btn
            text="Close"
            variant="text"
            @click="open = false"
          />
          <v-btn
            text="View pack"
            color="primary"
            :to="`/packs/${result.packId}`"
          />
        </template>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>
