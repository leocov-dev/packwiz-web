import {computed, getCurrentInstance, onBeforeUnmount, ref} from "vue"
import type {CompatResponse, CompatSummary} from "@/interfaces/compat.ts"
import {fetchCompat, requestCompatScan} from "@/services/compat.service.ts"
import {apiErrorMessage} from "@/services/utils.ts"
import {COMPAT_POLL_MS, COMPAT_STALLED_MESSAGE, abandonedStatus, isCompatActive, shouldContinueCompatPolling} from "@/lib/compat.ts"
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
  // Bumped whenever the state is replaced or polling restarts; async work started
  // under an older value must not touch the state when it lands.
  let generation = 0
  let rescanToken = 0

  const status = computed(() => summary.value?.status ?? "unknown")
  const isActive = computed(() => isCompatActive(status.value))

  /** Invalidates everything in flight (a superseded load can no longer clear `loading` itself). */
  const supersede = () => {
    generation++
    clearTimeout(timer)
    loading.value = false
    return generation
  }

  const stopPolling = () => {
    supersede()
  }

  const apply = (r: CompatResponse) => {
    compat.value = r
    summary.value = r.status === "unknown"
      ? null
      : {status: r.status, errors: r.counts?.errors ?? 0, warnings: r.counts?.warnings ?? 0}
  }

  /** Stops showing a busy status when polling is abandoned; counts are kept. */
  const abandonPolling = (message: string) => {
    if (summary.value) summary.value = {...summary.value, status: abandonedStatus(summary.value.status)}
    if (compat.value) compat.value = {...compat.value, status: abandonedStatus(compat.value.status)}
    error.value = message
  }

  const poll = (gen: number, startedAt: number, failures = 0) => {
    timer = setTimeout(async () => {
      if (gen !== generation) return
      try {
        const r = await fetchCompat(packId())
        if (gen !== generation) return
        apply(r)
        failures = 0
      } catch (e) {
        if (gen !== generation) return
        failures++
        if (shouldGiveUpPolling(failures)) {
          abandonPolling(apiErrorMessage(e, COMPAT_STALLED_MESSAGE))
          return
        }
        poll(gen, startedAt, failures)
        return
      }
      if (shouldContinueCompatPolling(status.value, Date.now() - startedAt)) {
        poll(gen, startedAt, failures)
      } else if (isActive.value) {
        abandonPolling(COMPAT_STALLED_MESSAGE)
      }
    }, COMPAT_POLL_MS)
  }

  const startPolling = () => {
    poll(supersede(), Date.now())
  }

  /** Loads the full result (e.g. when the dialog opens); resumes polling if a scan is active. */
  const load = async () => {
    const gen = supersede()
    loading.value = true
    error.value = ""
    try {
      const r = await fetchCompat(packId())
      if (gen !== generation) return
      apply(r)
      loading.value = false
      if (isCompatActive(r.status)) startPolling()
    } catch (e) {
      if (gen !== generation) return
      error.value = apiErrorMessage(e, "Failed to load compatibility results")
    } finally {
      if (gen === generation) loading.value = false
    }
  }

  /**
   * Starts (or joins) a scan and polls until it finishes. Resolves to the failure
   * message when the scan could not be started, so callers can surface it.
   */
  const rescan = async (): Promise<string | null> => {
    const token = ++rescanToken
    const gen = supersede()
    error.value = ""
    rescanning.value = true
    try {
      const r = await requestCompatScan(packId())
      if (gen !== generation) return null
      // counts stay those of the last result until the new one lands
      summary.value = {status: r.status, errors: summary.value?.errors ?? 0, warnings: summary.value?.warnings ?? 0}
      if (compat.value) compat.value = {...compat.value, status: r.status}
      if (isCompatActive(r.status)) {
        startPolling()
      } else {
        await load()
      }
      return null
    } catch (e) {
      if (gen !== generation) return null
      error.value = apiErrorMessage(e, "Failed to start the compatibility scan")
      return error.value
    } finally {
      if (token === rescanToken) rescanning.value = false
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
