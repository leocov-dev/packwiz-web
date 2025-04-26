<script setup lang="ts">
import {type Pack} from "@/interfaces/pack.ts";
import {addMod} from "@/services/mods.service.ts";
import type {AddModRequest} from "@/interfaces/requests.ts";

const {pack} = defineProps<{ pack: Pack }>()

const router = useRouter()

const error = ref(false)
const isValid = ref(false)
const loading = ref(false)

const data = ref({
  modSource: "",
  modUrl: "",
})

const rules = {
  sourceRequired: (value: string) => !!value || "Mod Source is required",
  urlRequired: (value: string) => !!value || "Mod Url is required",
}

const parseUrl = (url: string) => {
  if (url.includes("curseforge.com")) {
    data.value.modSource = "Curseforge"
  } else if (url.includes("modrinth.com")) {
    data.value.modSource = "Modrinth"
  } else {
    data.value.modSource = ""
  }
}

const submitForm = async () => {
  error.value = false
  loading.value = true

  let request: AddModRequest

  if (data.value.modSource === "Curseforge") {
    request = {
      curseforge: {
        url: data.value.modUrl,
      }
    }
  } else if (data.value.modSource === "Modrinth") {
    request = {
      modrinth: {
        url: data.value.modUrl,
      }
    }
  } else {
    error.value = true
    loading.value = false
    console.error(`Invalid mod source: ${data.value.modSource}`)
    return
  }

  try {
    await addMod(pack.slug, request)

    await router.push({path: `/packs/${pack.slug}`})
  } catch (e) {
    error.value = true
    console.error(e)
  } finally {
    loading.value = false
  }

}

const cancelForm = async () => {
  await router.push({path: `/packs/${pack.slug}`})
}


watch(
  () => data.value.modUrl,
  (newUrl: string) => {
    parseUrl(newUrl)
  },
)

</script>

<template>
  <div
    class="ma-6"
  >
    <v-alert
      v-if="error"
      class="mb-6"
      text="Error adding new mod..."
      type="error"
      icon="mdi-alert"
      closable
    />
    <v-card>
      <v-card-title>
        <h1 class="me-5">
          {{ pack.packData?.name || pack.slug }}
        </h1>
      </v-card-title>

      <v-card-subtitle>
        <h3>Add New Mod</h3>
      </v-card-subtitle>

      <v-form
        v-model="isValid"
        class="ma-6"
        @submit.prevent="submitForm"
      >
        <v-select
          v-model="data.modSource"
          :items="['Curseforge', 'Modrinth']"
          label="Mod Source"
          :rules="[rules.sourceRequired]"
        />

        <v-text-field
          v-model="data.modUrl"
          label="Mod URL"
          :rules="[rules.urlRequired]"
        />

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

