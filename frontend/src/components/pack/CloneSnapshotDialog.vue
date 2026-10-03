<script setup lang="ts">
import {nextTick} from "vue";
import type {PackResponse} from "@/interfaces/pack.ts";
import type {PackSnapshot} from "@/interfaces/snapshot.ts";
import SlugAndName from "@/components/forms/SlugAndName.vue";
import {cloneFromSnapshot} from "@/services/snapshots.service.ts";
import {apiErrorMessage} from "@/services/utils.ts";
import {useSnackbarStore} from "@/stores/snackbar.ts";

const open = defineModel<boolean>({required: true})

const {pack, snapshot} = defineProps<{
  pack: PackResponse
  snapshot: PackSnapshot
}>()

const snackbar = useSnackbarStore()

const slug = ref('')
const name = ref('')
const isValid = ref(false)
const isSubmitting = ref(false)
const error = ref('')

// SlugAndName derives the slug whenever the name changes, so the default name
// is set after the form is mounted and reset.
watch(open, async (isOpen) => {
  if (!isOpen) {
    return
  }
  error.value = ''
  slug.value = ''
  name.value = ''
  await nextTick()
  name.value = `${pack.name || pack.slug} copy`
})

const submit = async () => {
  if (!isValid.value || isSubmitting.value) {
    return
  }

  isSubmitting.value = true
  error.value = ''
  try {
    const created = await cloneFromSnapshot(pack.id, snapshot.id, {slug: slug.value, name: name.value})
    // stay on the history page; the toast links to the new pack
    snackbar.showSnackbar(`Created ${created.name || created.slug}`, 'success', 10000, {
      text: 'View pack',
      to: `/packs/${created.id}`,
    })
    open.value = false
  } catch (e) {
    error.value = apiErrorMessage(e, 'Failed to clone the pack.')
  } finally {
    isSubmitting.value = false
  }
}
</script>

<template>
  <v-dialog
    v-model="open"
    persistent
    max-width="600"
  >
    <v-card class="pa-3">
      <v-card-title>Clone snapshot #{{ snapshot.seq }}</v-card-title>
      <v-card-text>
        <p class="mb-6">
          Creates a new draft pack with the mods and settings from this snapshot.
          You become its owner. The original pack is not changed.
        </p>

        <v-form
          v-model="isValid"
          @submit.prevent="submit"
        >
          <SlugAndName
            v-model:slug="slug"
            v-model:name="name"
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
      <v-card-actions>
        <v-spacer />
        <v-btn
          text="Cancel"
          variant="text"
          :disabled="isSubmitting"
          @click="open = false"
        />
        <v-btn
          text="Clone"
          color="primary"
          :disabled="!isValid"
          :loading="isSubmitting"
          @click="submit"
        />
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>
