<script setup lang="ts">
import {type Pack} from "@/interfaces/pack.ts";
import {addMod, listMissingDependencies, getCurseforgeStatus} from "@/services/mods.service.ts";
import type {AddModRequest} from "@/interfaces/requests.ts";
import type {ModDependency, ModSearchResult} from "@/interfaces/pack.ts"
import MissingDependencies from "@/components/mods/MissingDependencies.vue";
import ModSearchResults from "@/components/mods/ModSearchResults.vue";
import {parseUrl as parseModSourceUrl, buildRequest as buildModRequest, modPageUrl, addedKey, normalizeUrl, searchFilterCaption, type ModSource} from "@/lib/mod-source.ts";
import {useModSearch, type SearchMode} from "@/composables/useModSearch.ts";
import {useSnackbarStore} from "@/stores/snackbar.ts";
import axios from "axios";

const {pack} = defineProps<{ pack: Pack }>()

const router = useRouter()
const snackbar = useSnackbarStore()

const error = ref(false)
const errorMsg = ref("")
const isValid = ref(false)
const loading = ref(false)
const depsLoading = ref(false)
const depsError = ref(false)
let lastDepsRequest: AddModRequest | undefined
const dependencies = ref<ModDependency[]>([])
const addModButtonText = computed(() =>
  dependencies.value.length > 0 ? "Add Mod and Dependencies" : "Add Mod"
)

const mode = ref<SearchMode>("modrinth")
const curseforgeAvailable = ref<boolean | null>(null)
const searchQuery = ref("")
const selectedProjectSlug = ref("")
const addingSlug = ref("")
const addedKeys = ref<string[]>([])
const modUrl = ref("")

const searchSource = computed<"modrinth" | "curseforge">(() =>
  mode.value === "curseforge" ? "curseforge" : "modrinth"
)
const trimmedUrl = computed(() => normalizeUrl(modUrl.value ?? ""))
const modSource = computed<ModSource>(() => parseModSourceUrl(trimmedUrl.value))

const {
  results: searchResults,
  loading: searchLoading,
  error: searchError,
  emptyState,
} = useModSearch({
  packId: () => pack.id,
  mcVersion: () => pack.mcVersion,
  mode,
  query: searchQuery,
  curseforgeAvailable,
  onReset: () => clearSelection(),
})

const filterCaption = computed(() =>
  searchFilterCaption(searchSource.value, pack.mcVersion, pack.loader)
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
  urlRequired: (value: string) => !!value?.trim() || "Mod Url is required",
  urlSupported: (value: string) => !value?.trim() || parseModSourceUrl(value.trim()) !== "" || "Enter a Modrinth, CurseForge or GitHub link, e.g. https://modrinth.com/mod/sodium",
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
  lastDepsRequest = request
  depsError.value = false
  depsLoading.value = true

  try {
    const deps = await listMissingDependencies(pack.id, request)

    if (seq === requestSeq) {
      dependencies.value = deps.missing ?? []
    }
  } catch (e) {
    if (seq === requestSeq) {
      depsError.value = true
    }
    console.error("Failed to check dependencies:", e)
  } finally {
    if (seq === requestSeq) {
      depsLoading.value = false
    }
  }
}

const errorMessage = (e: unknown): string =>
  axios.isAxiosError(e) ? (e.response?.data?.error || "Failed to add mod") : String(e)

const submitRequest = async (request: AddModRequest, title: string): Promise<boolean> => {
  error.value = false

  try {
    await addMod(pack.id, request)
    snackbar.showSnackbar(title === "mod" ? "Mod added" : `Added ${title}`, "success")
    return true
  } catch (e) {
    showError(errorMessage(e))
    console.error(errorMessage(e))
    return false
  }
}

const submitForm = async () => {
  if (mode.value !== "url") {
    return
  }

  const request = requestFor(modSource.value, trimmedUrl.value)

  if (request === undefined) {
    return
  }

  loading.value = true
  const ok = await submitRequest(request, "mod")
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

  const key = addedKey(searchSource.value, result.slug)

  addingSlug.value = result.slug
  const ok = await submitRequest(request, result.title)
  addingSlug.value = ""

  if (ok) {
    addedKeys.value.push(key)

    if (selectedProjectSlug.value === result.slug) {
      clearSelection()
    }
  }
}

const cancelForm = async () => {
  await router.push({path: `/packs/${pack.id}`})
}

const retryDependencies = () => {
  if (lastDepsRequest !== undefined && !depsLoading.value) {
    void checkForDependencies(lastDepsRequest, ++requestSeq)
  }
}

const clearSelection = () => {
  requestSeq++
  depsLoading.value = false
  depsError.value = false
  lastDepsRequest = undefined
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

onBeforeUnmount(() => {
  clearTimeout(debounceTimer)
})

watch(mode, () => {
  modUrl.value = ""
  error.value = false
  errorMsg.value = ""
  clearSelection()
})

watch(modUrl, (rawUrl: string | null) => {
  const newUrl = normalizeUrl(rawUrl ?? "")
  dependencies.value = []
  depsLoading.value = false
  depsError.value = false
  error.value = false

  if (debounceTimer) {
    clearTimeout(debounceTimer)
  }

  const seq = ++requestSeq

  if (!newUrl || parseModSourceUrl(newUrl) === "") {
    return
  }

  depsLoading.value = true

  debounceTimer = setTimeout(async () => {
    const request = requestFor(modSource.value, newUrl)
    if (request !== undefined) {
      await checkForDependencies(request, seq)
    } else if (seq === requestSeq) {
      depsLoading.value = false
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
        v-if="error && mode === 'url'"
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

      <div class="ma-6">
        <v-chip-group
          v-model="mode"
          class="mb-4"
          mandatory
          selected-class="text-primary"
          aria-label="Add mod method"
        >
          <v-chip
            value="modrinth"
            filter
            label
            variant="tonal"
            prepend-icon="mdi-magnify"
            text="Search Modrinth"
          />
          <v-chip
            value="curseforge"
            filter
            label
            variant="tonal"
            prepend-icon="mdi-magnify"
            :class="{'text-medium-emphasis': curseforgeAvailable === false}"
            text="Search CurseForge"
          />
          <v-chip
            value="url"
            filter
            label
            variant="tonal"
            prepend-icon="mdi-link-variant"
            text="Paste URL"
          />
        </v-chip-group>

        <v-form
          v-if="mode === 'url'"
          v-model="isValid"
          @submit.prevent="submitForm"
        >
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

          <v-alert
            v-if="depsError"
            type="warning"
            variant="tonal"
            density="compact"
            class="mb-4"
            text="Couldn't check dependencies. You can still add this mod, but required dependencies may be missing."
          >
            <template #append>
              <v-btn
                text="Retry"
                variant="text"
                size="small"
                :disabled="depsLoading"
                @click="retryDependencies"
              />
            </template>
          </v-alert>

          <div class="d-flex justify-end">
            <v-btn
              :text="addModButtonText"
              color="primary"
              type="submit"
              :disabled="loading || depsLoading || !isValid || !trimmedUrl"
            />
          </div>
        </v-form>

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
              :hint="filterCaption"
              persistent-hint
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
              :pack-loader="pack.loader"
              :installed-mods="pack.mods ?? []"
              :added-keys="addedKeys"
              :adding-slug="addingSlug"
              :selected-slug="selectedProjectSlug"
              :dependencies="dependencies"
              :error-message="error ? errorMsg : ''"
              :dependencies-failed="depsError"
              @select="selectSearchResult"
              @add="addResult"
              @retry-dependencies="retryDependencies"
            />

            <div
              v-else-if="!searchError && emptyState !== 'none'"
              class="text-medium-emphasis mb-4"
            >
              {{ emptyState === 'short-query' ? 'Type at least 2 characters to search' : 'No results' }}
            </div>
          </template>
        </div>
      </div>


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
