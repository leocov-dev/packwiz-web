import {computed, getCurrentInstance, onBeforeUnmount, ref} from "vue"
import type {CompatResponse, CompatSummary} from "@/interfaces/compat.ts"
import {fetchCompat, requestCompatScan} from "@/services/compat.service.ts"
import {apiErrorMessage} from "@/services/utils.ts"
import {COMPAT_POLL_MS, isCompatActive, shouldContinueCompatPolling} from "@/lib/compat.ts"
import {shouldGiveUpPolling} from "@/lib/update-checks.ts"

/**
 * Jar compatibility state of one pack. Starts from the summary the pack already
 * carries; the full result is loaded on demand (`load`). Polls while a scan is
 * queued or running (fixed interval, bounded duration) and stops on unmount.
 */
export function useCompat(packId: () => number, initial: () => CompatSummary | undefined) {
  const summary = ref<Pick<CompatSummary, "status" | "errors" | "warnings"> | null>(initial() ?? null)
  const compat = ref<CompatResponse | null>(null)
  const loading = ref(false)
  const rescanning = ref(false)
  const error = ref("")

  let timer: ReturnType<typeof setTimeout> | undefined
  let generation = 0

  const status = computed(() => summary.value?.status ?? "unknown")
  const isActive = computed(() => isCompatActive(status.value))

  const stopPolling = () => {
    generation++
    clearTimeout(timer)
  }

  const apply = (r: CompatResponse) => {
    compat.value = r
    summary.value = r.status === "unknown"
      ? null
      : {status: r.status, errors: r.counts?.errors ?? 0, warnings: r.counts?.warnings ?? 0}
  }

  const poll = (gen: number, startedAt: number, failures = 0) => {
    timer = setTimeout(async () => {
      if (gen !== generation) return
      try {
        apply(await fetchCompat(packId()))
        failures = 0
      } catch (e) {
        if (gen !== generation) return
        failures++
        if (shouldGiveUpPolling(failures)) {
          error.value = apiErrorMessage(e, "Failed to load compatibility results")
          return
        }
        poll(gen, startedAt, failures)
        return
      }
      if (gen !== generation) return
      if (shouldContinueCompatPolling(status.value, Date.now() - startedAt)) {
        poll(gen, startedAt, failures)
      } else if (isActive.value) {
        error.value = "The scan is taking too long; check back later"
      }
    }, COMPAT_POLL_MS)
  }

  const startPolling = () => {
    stopPolling()
    poll(generation, Date.now())
  }

  /** Loads the full result (e.g. when the dialog opens); resumes polling if a scan is active. */
  const load = async () => {
    const gen = ++generation
    clearTimeout(timer)
    loading.value = true
    error.value = ""
    try {
      const r = await fetchCompat(packId())
      if (gen !== generation) return
      apply(r)
      if (isCompatActive(r.status)) startPolling()
    } catch (e) {
      if (gen !== generation) return
      error.value = apiErrorMessage(e, "Failed to load compatibility results")
    } finally {
      if (gen === generation) loading.value = false
    }
  }

  /** Starts (or joins) a scan and polls until it finishes. */
  const rescan = async () => {
    error.value = ""
    rescanning.value = true
    try {
      const r = await requestCompatScan(packId())
      summary.value = {status: r.status, errors: summary.value?.errors ?? 0, warnings: summary.value?.warnings ?? 0}
      if (isCompatActive(r.status)) {
        startPolling()
      } else {
        await load()
      }
    } catch (e) {
      error.value = apiErrorMessage(e, "Failed to start the compatibility scan")
    } finally {
      rescanning.value = false
    }
  }

  /** Adopts a fresh summary from the pack (e.g. after a reload), resuming polling if needed. */
  const sync = (next: CompatSummary | undefined) => {
    stopPolling()
    compat.value = null
    error.value = ""
    summary.value = next ?? null
    if (isCompatActive(status.value)) startPolling()
  }

  /** Stops background polling (dialog-independent; also runs on unmount). */
  const stop = stopPolling

  if (isCompatActive(status.value)) startPolling()
  if (getCurrentInstance()) onBeforeUnmount(stopPolling)

  return {summary, compat, status, isActive, loading, rescanning, error, load, rescan, sync, stop}
}
