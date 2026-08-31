<script setup lang="ts">
import {type Pack} from "@/interfaces/pack.ts";
import {addMod, listMissingDependencies, searchModrinthMods} from "@/services/mods.service.ts";
import type {AddModRequest} from "@/interfaces/requests.ts";
import type {ModDependency, ModSearchResult} from "@/interfaces/pack.ts"
import MissingDependencies from "@/components/mods/MissingDependencies.vue";
import {parseUrl as parseModSourceUrl, buildRequest as buildModRequest, type ModSource} from "@/lib/mod-source.ts";
import axios from "axios";

const {pack} = defineProps<{ pack: Pack }>()

const router = useRouter()

const error = ref(false)
const errorMsg = ref("")
const isValid = ref(false)
const loading = ref(false)
const dependencies = ref<ModDependency[]>([])

const mode = ref<"url" | "search">("url")
const searchQuery = ref("")
const searchResults = ref<ModSearchResult[]>([])
const searchLoading = ref(false)
const selectedProjectSlug = ref("")

const data = ref({
  modSource: "",
  modUrl: "",
})

const rules = {
  sourceRequired: (value: string) => !!value || "Mod Source is required",
  urlRequired: (value: string) => !!value || "Mod Url is required",
}

const parseUrl = (url: string) => {
  data.value.modSource = parseModSourceUrl(url)
}

const checkForDependencies = async (seq: number) => {
  const request = buildRequest()

  if (request === undefined) {
    return
  }
  const deps = await listMissingDependencies(pack.id, request)

  if (seq === requestSeq) {
    dependencies.value = deps.missing
  }
}

const buildRequest = (): AddModRequest | undefined => {
  const result = buildModRequest(data.value.modSource as ModSource, data.value.modUrl)

  if ("request" in result) {
    return result.request
  }

  error.value = true
  errorMsg.value = result.error
  console.error(result.error)
}

const submitForm = async () => {
  error.value = false
  loading.value = true

  const request = buildRequest()

  if (request === undefined) {
    error.value = true
    loading.value = false
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

const selectSearchResult = (result: ModSearchResult) => {
  selectedProjectSlug.value = result.slug
  data.value.modUrl = `https://modrinth.com/mod/${result.slug}`
}


let debounceTimer: ReturnType<typeof setTimeout> | undefined
let requestSeq = 0

let searchDebounceTimer: ReturnType<typeof setTimeout> | undefined
let searchRequestSeq = 0

watch(searchQuery, (newQuery: string) => {
  searchResults.value = []

  if (searchDebounceTimer) {
    clearTimeout(searchDebounceTimer)
  }

  if (!newQuery || newQuery.length < 2) {
    searchLoading.value = false
    return
  }

  searchLoading.value = true
  const seq = ++searchRequestSeq

  searchDebounceTimer = setTimeout(async () => {
    try {
      const response = await searchModrinthMods(pack.id, newQuery, pack.mcVersion ? [pack.mcVersion] : undefined)
      if (seq === searchRequestSeq) {
        searchResults.value = response.results || []
      }
    } finally {
      if (seq === searchRequestSeq) {
        searchLoading.value = false
      }
    }
  }, 400)
})

watch(
  () => data.value.modUrl,
  (newUrl: string | null) => {
    dependencies.value = []
    error.value = false

    if (debounceTimer) {
      clearTimeout(debounceTimer)
    }

    if (!newUrl) {
      data.value.modSource = ""
      loading.value = false
      return
    }

    loading.value = true
    const seq = ++requestSeq

    debounceTimer = setTimeout(async () => {
      try {
        parseUrl(newUrl)
        await checkForDependencies(seq)
      } finally {
        if (seq === requestSeq) {
          loading.value = false
        }
      }
    }, 400)
  },
)

</script>

<template>
  <div
    class="ma-6"
  >
    <v-card>
      <v-card-title>
        <h1 class="me-5">
          {{ pack.name || pack.slug }}
        </h1>
      </v-card-title>

      <v-alert
        v-if="error"
        class="mb-6 ms-6 me-6"
        :text="'Error: ' + (errorMsg || 'failed to add new mod...')"
        type="error"
        icon="mdi-alert"
        closable
      />

      <v-card-subtitle>
        <h3>Add New Mod</h3>
      </v-card-subtitle>

      <v-form
        v-model="isValid"
        class="ma-6"
        @submit.prevent="submitForm"
      >
        <v-select
          v-show="false"
          v-model="data.modSource"
          :items="['Curseforge', 'Modrinth', 'GitHub']"
          label="Mod Source"
          :rules="[rules.sourceRequired]"
        />

        <v-btn-toggle
          v-model="mode"
          class="mb-4"
          mandatory
          density="comfortable"
          variant="outlined"
        >
          <v-btn
            value="url"
            text="Paste URL"
          />
          <v-btn
            value="search"
            text="Search Modrinth"
          />
        </v-btn-toggle>

        <v-text-field
          v-if="mode === 'url'"
          v-model="data.modUrl"
          label="Mod URL"
          :rules="[rules.urlRequired]"
          clearable
        />

        <div v-else>
          <v-text-field
            v-model="searchQuery"
            label="Search Modrinth"
            prepend-inner-icon="mdi-magnify"
            :loading="searchLoading"
            clearable
          />

          <v-list v-if="searchResults.length > 0">
            <v-list-item
              v-for="result in searchResults"
              :key="result.projectId"
              :active="result.slug === selectedProjectSlug"
              @click="selectSearchResult(result)"
            >
              <template #prepend>
                <v-avatar
                  v-if="result.iconUrl"
                  :image="result.iconUrl"
                />
                <v-icon
                  v-else
                  icon="mdi-puzzle-outline"
                />
              </template>
              <v-list-item-title>{{ result.title }}</v-list-item-title>
              <v-list-item-subtitle class="text-truncate">
                {{ result.description }}
              </v-list-item-subtitle>
            </v-list-item>
          </v-list>
        </div>

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
            :disabled="loading || !isValid || (mode === 'search' && !data.modUrl)"
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

