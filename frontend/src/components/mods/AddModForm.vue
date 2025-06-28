<script setup lang="ts">
import {type Pack} from "@/interfaces/pack.ts";
import {addMod, listMissingDependencies} from "@/services/mods.service.ts";
import type {AddModRequest} from "@/interfaces/requests.ts";
import type {ModDependency} from "@/interfaces/pack.ts"
import MissingDependencies from "@/components/mods/MissingDependencies.vue";
import axios from "axios";

const {pack} = defineProps<{ pack: Pack }>()

const router = useRouter()

const error = ref(false)
const errorMsg = ref("")
const isValid = ref(false)
const loading = ref(false)
const dependencies = ref<ModDependency[]>([])

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
  } else if (url.includes("github.com")) {
    data.value.modSource = "Github"
  } else {
    data.value.modSource = ""
  }
}

const checkForDependencies = async () => {
  const request = buildRequest()

  if (request === undefined) {
    return
  }
  const deps = await listMissingDependencies(pack.id, request)

  dependencies.value = deps.missing
}

const buildRequest = (): AddModRequest | undefined => {
  if (data.value.modSource === "Curseforge") {
    return {
      curseforge: {
        url: data.value.modUrl,
      }
    }
  } else if (data.value.modSource === "Modrinth") {
    return {
      modrinth: {
        url: data.value.modUrl,
      }
    }
  } else if (data.value.modSource === "Github") {
    return {
      github: {
        url: data.value.modUrl,
      }
    }
  }

  error.value = true
  console.error(`Invalid mod source: ${data.value.modSource}`)
}

const submitForm = async () => {
  error.value = false
  loading.value = true

  const request = buildRequest()

  if (request === undefined) {
    error.value = true
    loading.value = false
    errorMsg.value = `Invalid mod source: ${data.value.modSource}`
    console.warn(errorMsg.value)
    return
  }

  try {
    await addMod(pack.id, request)

    await router.push({path: `/packs/${pack.id}`})
  } catch (e) {
    error.value = true

    if (axios.isAxiosError(e)) {
      errorMsg.value = e.response?.data?.error || "Failed to add mod"
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


watch(
  () => data.value.modUrl,
  async (newUrl: string) => {
    error.value = false
    loading.value = true
    try {
      parseUrl(newUrl)
      await checkForDependencies()
    } finally {
      loading.value = false
    }
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
      :text="'Error: ' + (errorMsg || 'failed to add new mod...')"
      type="error"
      icon="mdi-alert"
      closable
    />
    <v-card>
      <v-card-title>
        <h1 class="me-5">
          {{ pack.name || pack.slug }}
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
          :items="['Curseforge', 'Modrinth', 'GitHub']"
          label="Mod Source"
          :rules="[rules.sourceRequired]"
        />

        <v-text-field
          v-model="data.modUrl"
          label="Mod URL"
          :rules="[rules.urlRequired]"
        />

        <MissingDependencies
          v-if="(dependencies || []).length > 0"
          class="mt-2 mb-6"
          :missing="dependencies"
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

