<script setup lang="ts">
import {type Pack} from "@/interfaces/pack.ts";
import {addMod, listMissingDependencies, searchModrinthMods, searchCurseforgeMods, getCurseforgeStatus} from "@/services/mods.service.ts";
import type {AddModRequest} from "@/interfaces/requests.ts";
import type {ModDependency, ModSearchResult} from "@/interfaces/pack.ts"
import MissingDependencies from "@/components/mods/MissingDependencies.vue";
import ModSearchResults from "@/components/mods/ModSearchResults.vue";
import {parseUrl as parseModSourceUrl, buildRequest as buildModRequest, modPageUrl, searchEmptyState, type ModSource} from "@/lib/mod-source.ts";
import {useSnackbarStore} from "@/stores/snackbar.ts";
import axios from "axios";

const {pack} = defineProps<{ pack: Pack }>()

const router = useRouter()
const snackbar = useSnackbarStore()

type Mode = "url" | "modrinth" | "curseforge"

const error = ref(false)
const errorMsg = ref("")
const isValid = ref(false)
const loading = ref(false)
const dependencies = ref<ModDependency[]>([])
const addModButtonText = computed(() =>
  dependencies.value.length > 0 ? "Add Mod and Dependencies" : "Add Mod"
)

const mode = ref<Mode>("modrinth")
const curseforgeAvailable = ref<boolean | null>(null)
const searchQuery = ref("")
const searchResults = ref<ModSearchResult[]>([])
const searchLoading = ref(false)
const searchError = ref("")
const hasSearched = ref(false)
const selectedProjectSlug = ref("")
const addingSlug = ref("")
const addedSlugs = ref<string[]>([])
const modUrl = ref("")

const searchSource = computed<"modrinth" | "curseforge">(() =>
  mode.value === "curseforge" ? "curseforge" : "modrinth"
)
const modSource = computed<ModSource>(() => parseModSourceUrl(modUrl.value))
const emptyState = computed(() =>
  searchEmptyState(searchQuery.value, searchLoading.value, searchResults.value.length, hasSearched.value)
)

onMounted(async () => {
  try {
    const status = await getCurseforgeStatus(pack.id)
    curseforgeAvailable.value = status.available
  } catch (e) {
    console.error("Failed to check CurseForge status:", e)
    curseforgeAvailable.value = false
  }
})

const rules = {
  urlRequired: (value: string) => !!value || "Mod Url is required",
  urlSupported: (value: string) => !value || parseModSourceUrl(value) !== "" || "URL must be a Modrinth, CurseForge or GitHub link",
}

let requestSeq = 0

const showError = (msg: string) => {
  error.value = true
  errorMsg.value = msg
}

const requestFor = (source: ModSource, url: string): AddModRequest | undefined => {
  const result = buildModRequest(source, url)

  if ("request" in result) {
    return result.request
  }

  showError(result.error)
}

const resultRequest = (result: ModSearchResult): AddModRequest | undefined =>
  requestFor(
    searchSource.value === "curseforge" ? "Curseforge" : "Modrinth",
    modPageUrl(searchSource.value, result.slug),
  )

const checkForDependencies = async (request: AddModRequest, seq: number) => {
  try {
    const deps = await listMissingDependencies(pack.id, request)

    if (seq === requestSeq) {
      dependencies.value = deps.missing
    }
  } catch (e) {
    console.error("Failed to check dependencies:", e)
  }
}

const errorMessage = (e: unknown): string =>
  axios.isAxiosError(e) ? (e.response?.data?.error || "Failed to add mod") : String(e)

const submitRequest = async (request: AddModRequest, title: string): Promise<boolean> => {
  error.value = false

  try {
    await addMod(pack.id, request)
    snackbar.showSnackbar(`Added ${title}`, "success")
    return true
  } catch (e) {
    showError(errorMessage(e))
    console.error(errorMessage(e))
    return false
  }
}

const submitForm = async () => {
  const request = requestFor(modSource.value, modUrl.value)

  if (request === undefined) {
    return
  }

  loading.value = true
  const url = modUrl.value
  const ok = await submitRequest(request, url)
  loading.value = false

  if (ok) {
    modUrl.value = ""
  }
}

const addResult = async (result: ModSearchResult) => {
  const request = resultRequest(result)

  if (request === undefined) {
    return
  }

  addingSlug.value = result.slug
  const ok = await submitRequest(request, result.title)
  addingSlug.value = ""

  if (ok) {
    addedSlugs.value.push(result.slug)
    clearSelection()
  }
}

const cancelForm = async () => {
  await router.push({path: `/packs/${pack.id}`})
}

const clearSelection = () => {
  requestSeq++
  selectedProjectSlug.value = ""
  dependencies.value = []
}

const selectSearchResult = (result: ModSearchResult) => {
  if (selectedProjectSlug.value === result.slug) {
    clearSelection()
    return
  }

  clearSelection()
  selectedProjectSlug.value = result.slug
  error.value = false

  const request = resultRequest(result)

  if (request !== undefined) {
    void checkForDependencies(request, requestSeq)
  }
}

let debounceTimer: ReturnType<typeof setTimeout> | undefined

let searchDebounceTimer: ReturnType<typeof setTimeout> | undefined
let searchRequestSeq = 0

const executeSearch = (query: string, searchMode: "modrinth" | "curseforge") => {
  searchResults.value = []
  searchError.value = ""
  hasSearched.value = false
  clearSelection()

  if (searchDebounceTimer) {
    clearTimeout(searchDebounceTimer)
  }

  const seq = ++searchRequestSeq

  if (searchMode === "curseforge" && !curseforgeAvailable.value) {
    searchLoading.value = false
    return
  }

  if (!query || query.length < 2) {
    searchLoading.value = false
    return
  }

  searchLoading.value = true

  searchDebounceTimer = setTimeout(async () => {
    try {
      const versions = pack.mcVersion ? [pack.mcVersion] : undefined
      const response = searchMode === "curseforge"
        ? await searchCurseforgeMods(pack.id, query, versions)
        : await searchModrinthMods(pack.id, query, versions)

      if (seq === searchRequestSeq) {
        searchResults.value = response.results || []
      }
    } catch (e) {
      if (seq === searchRequestSeq) {
        searchResults.value = []
        searchError.value = axios.isAxiosError(e)
          ? (e.response?.data?.error || "Search failed")
          : "Search failed"
      }
      console.error("Search failed:", e)
    } finally {
      if (seq === searchRequestSeq) {
        searchLoading.value = false
        hasSearched.value = true
      }
    }
  }, 400)
}

watch(searchQuery, (newQuery: string | null) => {
  if (mode.value !== "url") {
    executeSearch(newQuery ?? "", mode.value)
  }
})

watch(mode, (newMode) => {
  modUrl.value = ""
  error.value = false
  errorMsg.value = ""
  clearSelection()
  if (newMode === "url") {
    searchResults.value = []
    searchError.value = ""
    hasSearched.value = false
    searchRequestSeq++
    searchLoading.value = false
  } else {
    executeSearch(searchQuery.value, newMode)
  }
})

watch(modUrl, (newUrl: string | null) => {
  dependencies.value = []
  error.value = false

  if (debounceTimer) {
    clearTimeout(debounceTimer)
  }

  const seq = ++requestSeq

  if (!newUrl || parseModSourceUrl(newUrl) === "") {
    return
  }

  debounceTimer = setTimeout(async () => {
    const request = requestFor(modSource.value, newUrl)
    if (request !== undefined) {
      await checkForDependencies(request, seq)
    }
  }, 400)
})

</script>

<template>
  <div
    class="ma-6"
  >
    <v-card>
      <v-card-title class="d-flex align-center">
        <v-btn
          icon="mdi-arrow-left"
          variant="text"
          class="me-3"
          :disabled="loading"
          @click="cancelForm"
        />
        <h1 class="me-5">
          {{ pack.name || pack.slug }}
        </h1>
        <v-spacer />
        <v-btn
          text="Back to pack"
          variant="text"
          :disabled="loading"
          @click="cancelForm"
        />
      </v-card-title>

      <v-alert
        v-if="error"
        v-model="error"
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
        <v-btn-toggle
          v-model="mode"
          class="mb-4"
          mandatory
          density="comfortable"
          variant="outlined"
        >
          <v-btn
            value="modrinth"
            text="Search Modrinth"
          />
          <v-btn
            value="curseforge"
            text="Search CurseForge"
          />
          <v-btn
            value="url"
            text="Paste URL"
          />
        </v-btn-toggle>

        <template v-if="mode === 'url'">
          <v-text-field
            v-model="modUrl"
            label="Mod URL"
            :rules="[rules.urlRequired, rules.urlSupported]"
            clearable
          />

          <MissingDependencies
            v-if="dependencies.length > 0"
            class="mb-4"
            :missing="dependencies"
          />

          <div class="d-flex justify-end">
            <v-btn
              :text="addModButtonText"
              color="primary"
              type="submit"
              :disabled="loading || !isValid || !modUrl"
            />
          </div>
        </template>

        <div v-else>
          <v-alert
            v-if="mode === 'curseforge' && curseforgeAvailable === false"
            type="warning"
            variant="tonal"
            icon="mdi-key-alert"
            title="CurseForge API Key Required"
            text="CurseForge search requires a CurseForge API key. Please configure the PWW_CF_API_KEY environment variable on the server to enable CurseForge search."
            class="mb-4"
          />

          <template v-else>
            <v-text-field
              v-model="searchQuery"
              :label="mode === 'curseforge' ? 'Search CurseForge' : 'Search Modrinth'"
              prepend-inner-icon="mdi-magnify"
              :loading="searchLoading"
              clearable
            />

            <v-alert
              v-if="searchError"
              type="error"
              variant="tonal"
              density="compact"
              class="mb-4"
              :text="searchError"
            />

            <ModSearchResults
              v-if="searchResults.length > 0"
              :results="searchResults"
              :source="searchSource"
              :installed-mods="pack.mods"
              :added-slugs="addedSlugs"
              :adding-slug="addingSlug"
              :selected-slug="selectedProjectSlug"
              :dependencies="dependencies"
              @select="selectSearchResult"
              @add="addResult"
            />

            <div
              v-else-if="!searchError && emptyState !== 'none'"
              class="text-medium-emphasis mb-4"
            >
              {{ emptyState === 'short-query' ? 'Type at least 2 characters to search' : 'No results' }}
            </div>
          </template>
        </div>

        <div class="d-flex justify-end mt-2">
          <v-btn
            text="Done"
            variant="text"
            :disabled="loading"
            @click="cancelForm"
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
