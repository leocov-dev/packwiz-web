<script setup lang="ts">
import type {PackResponse} from "@/interfaces/pack.ts"
import type {RehashRequest} from "@/interfaces/requests.ts"
import {rehashAllMods} from "@/services/packs.service.ts"
import {useSnackbarStore} from "@/stores/snackbar.ts"

const {pack} = defineProps<{ pack: PackResponse }>()
const model = defineModel<boolean>({required: true})
const emit = defineEmits(["rehashed"])

const snackbar = useSnackbarStore()

const format = ref<RehashRequest["format"]>("sha256")
const loading = ref(false)

const submit = async () => {
  loading.value = true
  try {
    await rehashAllMods(pack.id, {format: format.value})
    model.value = false
    emit("rehashed")
  } catch (e) {
    console.error(e)
    snackbar.showSnackbar("Failed to rehash mods", "error", 4000)
  } finally {
    loading.value = false
  }
}

const cancel = () => {
  model.value = false
}
</script>

<template>
  <v-dialog
    v-model="model"
    persistent
    max-width="400"
  >
    <v-card class="pa-3">
      <v-card-title>Rehash All Mods</v-card-title>
      <v-card-text>
        <v-select
          v-model="format"
          :items="[
            {title: 'SHA-1', value: 'sha1'},
            {title: 'SHA-256', value: 'sha256'},
            {title: 'SHA-512', value: 'sha512'},
          ]"
          label="Hash format"
        />
      </v-card-text>
      <v-card-actions>
        <v-spacer />
        <v-btn
          class="me-3"
          color="primary"
          variant="tonal"
          :disabled="loading"
          @click="cancel"
        >
          Cancel
        </v-btn>
        <v-btn
          color="primary"
          variant="flat"
          :loading="loading"
          :disabled="loading"
          @click="submit"
        >
          Rehash
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>
