<script setup lang="ts">
import {type Pack, type Mod} from "@/interfaces/pack.ts";
import {changeModSide, pinMod, unpinMod, updateModFromSource} from "@/services/mods.service.ts";
import axios from "axios";

const {pack, mod} = defineProps<{ pack: Pack, mod: Mod }>()

const router = useRouter()

const error = ref(false)
const errorMsg = ref("")
const isValid = ref(false)
const loading = ref(false)

const data = ref({
  side: mod.side,
  pinned: mod.pinned,
})

const submitForm = async () => {
  error.value = false
  loading.value = true

  try {
    if (data.value.side !== mod.side) {
      await changeModSide(pack.id, mod.id, {side: data.value.side})
    }

    if (data.value.pinned !== mod.pinned) {
      if (data.value.pinned) {
        await pinMod(pack.id, mod.id)
      } else {
        await unpinMod(pack.id, mod.id)
      }
    }

    await router.push({path: `/packs/${pack.id}`})
  } catch (e) {
    error.value = true

    if (axios.isAxiosError(e)) {
      errorMsg.value = e.response?.data?.error || "Failed to update mod"
    } else {
      errorMsg.value = String(e)
    }

    console.error(errorMsg.value)
  } finally {
    loading.value = false
  }
}

const updateFromSource = async () => {
  error.value = false
  loading.value = true

  try {
    await updateModFromSource(pack.id, mod.id)
    await router.push({path: `/packs/${pack.id}`})
  } catch (e) {
    error.value = true

    if (axios.isAxiosError(e)) {
      errorMsg.value = e.response?.data?.error || "Failed to update mod from source"
    } else {
      errorMsg.value = String(e)
    }

    console.error(errorMsg.value)
  } finally {
    loading.value = false
  }
}

const cancelForm = async () => {
  await router.push({path: `/packs/${pack.id}`})
}
</script>

<template>
  <div
    class="ma-6"
  >
    <v-alert
      v-if="error"
      class="mb-6"
      :text="'Error: ' + (errorMsg || 'failed to edit mod...')"
      type="error"
      icon="mdi-alert"
      closable
    />
    <v-card>
      <v-card-title class="d-flex align-baseline">
        <h1 class="me-5">
          {{ pack.name || pack.slug }}
        </h1>
        <h2 class="me-5">
          {{ mod.name || mod.slug }}
        </h2>
      </v-card-title>

      <v-card-subtitle>
        <h3>Edit Mod</h3>
      </v-card-subtitle>

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

        <div class="d-flex justify-space-between mt-6">
          <v-btn
            text="Update from source"
            variant="outlined"
            :disabled="loading"
            @click="updateFromSource"
          />

          <div class="d-flex justify-end">
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
              :disabled="loading"
            />
          </div>
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
  </div>
</template>
