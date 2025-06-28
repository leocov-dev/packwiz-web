<script setup lang="ts">
import {type Pack, type Mod} from "@/interfaces/pack.ts";

const {pack, mod} = defineProps<{ pack: Pack, mod: Mod }>()

const router = useRouter()

const error = ref(false)
const isValid = ref(false)
const loading = ref(false)




const submitForm = async () => {
  error.value = false
  loading.value = true

  try {
    // await addMod(pack.id, request)

    await router.push({path: `/packs/${pack.id}`})
  } catch (e) {
    error.value = true
    console.error(e)
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
      text="Error editing mod..."
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
        <div class="d-flex justify-end">
          <v-btn
            text="Cancel"
            :disabled="loading"
            class="me-6"
            @click="cancelForm"
          />
          <v-btn
            text="Add Mod"
            color="primary"
            type="submit"
            :disabled="loading || !isValid"
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
  </div>
</template>
