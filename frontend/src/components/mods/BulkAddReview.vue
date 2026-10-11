<script setup lang="ts">
import type {BulkAddItem, ModDependency, ModSearchResult, Pack} from "@/interfaces/pack.ts"
import ModSearchResults from "@/components/mods/ModSearchResults.vue"
import {listMissingDependencies, resolveBulkItem} from "@/services/mods.service.ts"
import {apiErrorMessage} from "@/services/utils.ts"
import {addedKey, buildRequest, modPageUrl, searchFilterCaption} from "@/lib/mod-source.ts"
import {initialReviewSource} from "@/lib/bulk-add.ts"
import {useModSearch, type SearchMode} from "@/composables/useModSearch.ts"

const {pack, runId, item, curseforgeAvailable} = defineProps<{
  pack: Pack
  runId: number
  item: BulkAddItem
  curseforgeAvailable: boolean
}>()

const emit = defineEmits<{
  resolved: [item: BulkAddItem]
}>()

const mode = ref<SearchMode>(initialReviewSource(item, curseforgeAvailable))
const query = ref("")
const cfAvailable = computed<boolean | null>(() => curseforgeAvailable)

const selectedSlug = ref("")
const dependencies = ref<ModDependency[]>([])
const depsFailed = ref(false)
const addingSlug = ref("")
const rejecting = ref(false)
const errorMsg = ref("")
let depsSeq = 0

const source = computed<"modrinth" | "curseforge">(() => mode.value === "curseforge" ? "curseforge" : "modrinth")

const clearSelection = () => {
  depsSeq++
  selectedSlug.value = ""
  dependencies.value = []
  depsFailed.value = false
  errorMsg.value = ""
}

const {results, loading, error: searchError, emptyState} = useModSearch({
  packId: () => pack.id,
  mcVersion: () => pack.mcVersion,
  mode,
  query,
  curseforgeAvailable: cfAvailable,
  onReset: clearSelection,
})

const filterCaption = computed(() => searchFilterCaption(source.value, pack.mcVersion, pack.loader))

onMounted(() => {
  // set after mount so the search watcher runs with the line as the query
  query.value = item.line
})

const checkDependencies = async (result: ModSearchResult) => {
  const built = buildRequest(source.value === "curseforge" ? "Curseforge" : "Modrinth", modPageUrl(source.value, result.slug))
  if (!("request" in built)) return
  const seq = ++depsSeq
  try {
    const deps = await listMissingDependencies(pack.id, built.request)
    if (seq === depsSeq) dependencies.value = deps.missing ?? []
  } catch (e) {
    if (seq === depsSeq) depsFailed.value = true
    console.error("Failed to check dependencies:", e)
  }
}

const select = (result: ModSearchResult) => {
  const same = selectedSlug.value === result.slug
  clearSelection()
  if (same) return
  selectedSlug.value = result.slug
  void checkDependencies(result)
}

const retryDependencies = () => {
  const result = results.value.find(r => r.slug === selectedSlug.value)
  if (result) {
    depsFailed.value = false
    void checkDependencies(result)
  }
}

const pick = async (result: ModSearchResult) => {
  errorMsg.value = ""
  addingSlug.value = result.slug
  try {
    const updated = await resolveBulkItem(pack.id, runId, item.id, {
      action: "pick",
      source: source.value,
      slug: result.slug,
      projectId: result.projectId,
    })
    emit("resolved", updated)
  } catch (e) {
    errorMsg.value = apiErrorMessage(e, "Failed to add mod")
  } finally {
    addingSlug.value = ""
  }
}

const reject = async () => {
  errorMsg.value = ""
  rejecting.value = true
  try {
    emit("resolved", await resolveBulkItem(pack.id, runId, item.id, {action: "reject"}))
  } catch (e) {
    errorMsg.value = apiErrorMessage(e, "Failed to reject line")
  } finally {
    rejecting.value = false
  }
}
</script>

<template>
  <div>
    <div class="d-flex align-center flex-wrap ga-2 mb-2">
      <v-chip-group
        v-model="mode"
        mandatory
        selected-class="text-primary"
        aria-label="Search source"
      >
        <v-chip
          value="modrinth"
          filter
          label
          variant="tonal"
          prepend-icon="mdi-magnify"
          text="Modrinth"
        />
        <v-chip
          v-if="curseforgeAvailable"
          value="curseforge"
          filter
          label
          variant="tonal"
          prepend-icon="mdi-magnify"
          text="CurseForge"
        />
      </v-chip-group>
      <v-spacer />
      <v-btn
        text="Reject line"
        prepend-icon="mdi-close"
        variant="text"
        :loading="rejecting"
        :disabled="addingSlug !== ''"
        @click="reject"
      />
    </div>

    <v-text-field
      v-model="query"
      label="Search"
      prepend-inner-icon="mdi-magnify"
      :loading="loading"
      :hint="filterCaption"
      persistent-hint
      clearable
      class="mb-2"
    />

    <v-alert
      v-if="searchError"
      type="error"
      variant="tonal"
      density="compact"
      class="mb-4"
      :text="searchError"
    />

    <v-alert
      v-if="errorMsg && !selectedSlug"
      type="error"
      variant="tonal"
      density="compact"
      class="mb-4"
      :text="errorMsg"
    />

    <ModSearchResults
      v-if="results.length > 0"
      :results="results"
      :source="source"
      :pack-loader="pack.loader"
      :installed-mods="pack.mods ?? []"
      :added-keys="item.status === 'picked' ? [addedKey(source, item.slug)] : []"
      :adding-slug="addingSlug"
      :selected-slug="selectedSlug"
      :dependencies="dependencies"
      :error-message="errorMsg"
      :dependencies-failed="depsFailed"
      @select="select"
      @add="pick"
      @retry-dependencies="retryDependencies"
    />

    <div
      v-else-if="!searchError && emptyState !== 'none'"
      class="text-medium-emphasis mb-4"
    >
      {{ emptyState === 'short-query' ? 'Type at least 2 characters to search' : 'No results' }}
    </div>
  </div>
</template>
