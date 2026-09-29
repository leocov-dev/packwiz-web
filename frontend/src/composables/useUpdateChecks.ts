import {computed, getCurrentInstance, onBeforeUnmount, onMounted, ref} from "vue"
import type {UpdateCheckStatus} from "@/interfaces/pack.ts"
import {checkForUpdates, fetchUpdateChecks} from "@/services/packs.service.ts"
import {apiErrorMessage} from "@/services/utils.ts"
import {
  buildResultsMap,
  isActiveStatus,
  nextPollDelay,
  shouldContinuePolling,
  type UpdateChecksMap,
} from "@/lib/update-checks.ts"
import type {UpdateCheckItem} from "@/interfaces/pack.ts"

/**
 * Update-check state of one pack: loads the stored state on mount, can start a
 * check job and polls until it finishes (stops on unmount or after a max duration).
 */
export function useUpdateChecks(packId: () => number) {
  const status = ref<UpdateCheckStatus>("idle")
  const results = ref<UpdateChecksMap>(new Map<number, UpdateCheckItem>())
  const availableCount = ref(0)
  const checkedAt = ref<string | null>(null)
  const error = ref("")

  let timer: ReturnType<typeof setTimeout> | undefined
  let generation = 0

  const isChecking = computed(() => isActiveStatus(status.value))

  const stopPolling = () => {
    generation++
    clearTimeout(timer)
  }

  const apply = (r: Awaited<ReturnType<typeof fetchUpdateChecks>>) => {
    status.value = r.status
    results.value = buildResultsMap(r.results)
    availableCount.value = r.availableCount
    checkedAt.value = r.checkedAt ?? null
    error.value = r.status === "failed" ? (r.error || "Update check failed") : ""
  }

  const poll = (gen: number, startedAt: number, attempt: number) => {
    timer = setTimeout(async () => {
      if (gen !== generation) return
      try {
        const r = await fetchUpdateChecks(packId())
        if (gen !== generation) return
        apply(r)
      } catch (e) {
        if (gen !== generation) return
        status.value = "failed"
        error.value = apiErrorMessage(e, "Failed to load update check")
        return
      }
      const elapsed = Date.now() - startedAt
      if (shouldContinuePolling(status.value, elapsed)) {
        poll(gen, startedAt, attempt + 1)
      } else if (isActiveStatus(status.value)) {
        status.value = "failed"
        error.value = "Update check is taking too long; try again later"
      }
    }, nextPollDelay(attempt))
  }

  const startPolling = () => {
    stopPolling()
    poll(generation, Date.now(), 0)
  }

  /** Loads stored state once; resumes polling if a check is still running. */
  const refresh = async () => {
    const gen = ++generation
    try {
      const r = await fetchUpdateChecks(packId())
      if (gen !== generation) return
      apply(r)
      if (isActiveStatus(r.status)) startPolling()
    } catch (e) {
      if (gen !== generation) return
      error.value = apiErrorMessage(e, "Failed to load update check")
    }
  }

  /** Starts a check (or joins a running one) and polls until it completes. */
  const start = async (force = false) => {
    error.value = ""
    try {
      const job = await checkForUpdates(packId(), force)
      status.value = job.status
      if (isActiveStatus(job.status)) {
        startPolling()
      } else {
        // cached result within the server TTL
        await refresh()
      }
    } catch (e) {
      status.value = "failed"
      error.value = apiErrorMessage(e, "Failed to start update check")
    }
  }

  /** Drops local results (e.g. after Update All changed the mods). */
  const reset = () => {
    stopPolling()
    status.value = "idle"
    results.value = new Map()
    availableCount.value = 0
    checkedAt.value = null
    error.value = ""
  }

  if (getCurrentInstance()) {
    onMounted(refresh)
    onBeforeUnmount(stopPolling)
  }

  return {status, results, availableCount, checkedAt, error, isChecking, start, refresh, reset}
}
