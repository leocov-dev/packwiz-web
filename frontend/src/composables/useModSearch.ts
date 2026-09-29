import {onScopeDispose, ref, computed, watch, type Ref} from "vue";
import axios from "axios";
import {searchCurseforgeMods, searchModrinthMods} from "@/services/mods.service.ts";
import type {ModSearchResult} from "@/interfaces/pack.ts";
import {normalizeSearchQuery, searchEmptyState} from "@/lib/mod-source.ts";

export type SearchMode = "url" | "modrinth" | "curseforge"

export const SEARCH_DEBOUNCE_MS = 400

export interface ModSearchOptions {
  packId: () => number
  mcVersion: () => string | undefined
  mode: Ref<SearchMode>
  query: Ref<string | null | undefined>
  /** null while the status is still loading */
  curseforgeAvailable: Ref<boolean | null>
  /** called whenever results are reset (new query / mode change) */
  onReset?: () => void
}

/**
 * Debounced, race-safe mod search. Watches `mode` and `query`; stale responses
 * (superseded by a newer query or a mode switch) are discarded.
 */
export function useModSearch(options: ModSearchOptions) {
  const {packId, mcVersion, mode, query, curseforgeAvailable, onReset} = options

  const results = ref<ModSearchResult[]>([])
  const loading = ref(false)
  const error = ref("")
  const hasSearched = ref(false)

  let timer: ReturnType<typeof setTimeout> | undefined
  let seq = 0

  const emptyState = computed(() =>
    searchEmptyState(query.value ?? "", loading.value, results.value.length, hasSearched.value)
  )

  const clearPending = () => {
    seq++
    clearTimeout(timer)
    loading.value = false
  }

  const reset = () => {
    results.value = []
    error.value = ""
    hasSearched.value = false
    onReset?.()
  }

  const run = (searchMode: "modrinth" | "curseforge", text: string) => {
    reset()
    clearTimeout(timer)
    const current = ++seq

    if (searchMode === "curseforge" && !curseforgeAvailable.value) {
      loading.value = false
      return
    }

    const normalized = normalizeSearchQuery(text)

    if (normalized.length < 2) {
      loading.value = false
      return
    }

    loading.value = true

    timer = setTimeout(async () => {
      try {
        const mc = mcVersion()
        const versions = mc ? [mc] : undefined
        const response = searchMode === "curseforge"
          ? await searchCurseforgeMods(packId(), normalized, versions)
          : await searchModrinthMods(packId(), normalized, versions)

        if (current === seq) {
          results.value = response.results || []
        }
      } catch (e) {
        if (current === seq) {
          results.value = []
          error.value = axios.isAxiosError(e)
            ? (e.response?.data?.error || "Search failed")
            : "Search failed"
        }
        console.error("Search failed:", e)
      } finally {
        if (current === seq) {
          loading.value = false
          hasSearched.value = true
        }
      }
    }, SEARCH_DEBOUNCE_MS)
  }

  watch(query, (text) => {
    if (mode.value !== "url") {
      run(mode.value, text ?? "")
    }
  })

  watch(mode, (newMode) => {
    if (newMode === "url") {
      clearPending()
      reset()
    } else {
      run(newMode, query.value ?? "")
    }
  })

  onScopeDispose(() => {
    seq++
    clearTimeout(timer)
  })

  return {results, loading, error, hasSearched, emptyState}
}
