import {getCurrentInstance, onBeforeUnmount, onMounted, ref} from "vue"
import type {BulkAddItem, BulkAddRun} from "@/interfaces/pack.ts"
import {fetchBulkAdd, fetchLatestBulkAdd, startBulkAdd} from "@/services/mods.service.ts"
import {apiErrorMessage} from "@/services/utils.ts"
import {nextPollDelay, shouldGiveUpPolling} from "@/lib/update-checks.ts"
import {BULK_POLL_MAX_DURATION_MS, isActiveBulkStatus, replaceItem} from "@/lib/bulk-add.ts"

/**
 * Bulk add state of one pack: loads the latest run on mount (resuming polling
 * if it is still running), starts a run and polls until it finishes.
 */
export function useBulkAdd(packId: () => number) {
  const run = ref<BulkAddRun | null>(null)
  const loading = ref(true)
  const starting = ref(false)
  const error = ref("")

  let timer: ReturnType<typeof setTimeout> | undefined
  let generation = 0

  const stopPolling = () => {
    generation++
    clearTimeout(timer)
  }

  const poll = (gen: number, runId: number, startedAt: number, attempt: number, failures = 0) => {
    timer = setTimeout(async () => {
      if (gen !== generation) return
      try {
        const r = await fetchBulkAdd(packId(), runId)
        if (gen !== generation) return
        run.value = r
        failures = 0
      } catch (e) {
        if (gen !== generation) return
        failures++
        if (shouldGiveUpPolling(failures)) {
          error.value = apiErrorMessage(e, "Failed to load the bulk add")
          return
        }
        poll(gen, runId, startedAt, attempt + 1, failures)
        return
      }
      if (!isActiveBulkStatus(run.value?.status)) return
      if (Date.now() - startedAt < BULK_POLL_MAX_DURATION_MS) {
        poll(gen, runId, startedAt, attempt + 1, failures)
      } else {
        error.value = "The bulk add is taking too long; reload the page to check again"
      }
    }, nextPollDelay(attempt))
  }

  const startPolling = (runId: number) => {
    stopPolling()
    poll(generation, runId, Date.now(), 0)
  }

  /** Loads the latest run; resumes polling if it is still running. */
  const refresh = async () => {
    const gen = ++generation
    loading.value = true
    try {
      const r = await fetchLatestBulkAdd(packId())
      if (gen !== generation) return
      run.value = r
      if (r && isActiveBulkStatus(r.status)) startPolling(r.id)
    } catch (e) {
      if (gen !== generation) return
      error.value = apiErrorMessage(e, "Failed to load the last bulk add")
    } finally {
      if (gen === generation) loading.value = false
    }
  }

  /** Starts a run for the pasted list. Resolves true when it was accepted. */
  const start = async (text: string): Promise<boolean> => {
    error.value = ""
    starting.value = true
    try {
      const r = await startBulkAdd(packId(), {text})
      run.value = r
      if (isActiveBulkStatus(r.status)) startPolling(r.id)
      return true
    } catch (e) {
      error.value = apiErrorMessage(e, "Failed to start the bulk add")
      return false
    } finally {
      starting.value = false
    }
  }

  /** Applies one reviewed line to the local run. */
  const updateItem = (item: BulkAddItem) => {
    if (run.value) {
      run.value = {...run.value, items: replaceItem(run.value.items, item)}
    }
  }

  /** Forgets the shown run (to paste a new list); the server keeps it. */
  const clear = () => {
    stopPolling()
    run.value = null
    error.value = ""
  }

  if (getCurrentInstance()) {
    onMounted(refresh)
    onBeforeUnmount(stopPolling)
  }

  return {run, loading, starting, error, start, refresh, updateItem, clear}
}
