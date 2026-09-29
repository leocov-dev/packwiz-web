import {afterEach, beforeEach, describe, expect, it, vi} from "vitest"
import {effectScope, nextTick, ref} from "vue"
import {useModSearch, SEARCH_DEBOUNCE_MS, type SearchMode} from "./useModSearch.ts"
import {searchCurseforgeMods, searchModrinthMods} from "@/services/mods.service.ts"
import type {ModSearchResponse, ModSearchResult} from "@/interfaces/pack.ts"

vi.mock("@/services/mods.service.ts", () => ({
  searchModrinthMods: vi.fn(),
  searchCurseforgeMods: vi.fn(),
}))

const modrinth = vi.mocked(searchModrinthMods)
const curseforge = vi.mocked(searchCurseforgeMods)

const response = (...slugs: string[]): ModSearchResponse => ({
  results: slugs.map((slug) => ({slug, title: slug, projectId: slug} as ModSearchResult)),
})

const deferred = <T>() => {
  let resolve!: (v: T) => void
  const promise = new Promise<T>((r) => { resolve = r })
  return {promise, resolve}
}

const setup = (initial: SearchMode = "modrinth", cf: boolean | null = true) => {
  const mode = ref<SearchMode>(initial)
  const query = ref<string | null>("")
  const curseforgeAvailable = ref<boolean | null>(cf)
  const scope = effectScope()
  const search = scope.run(() =>
    useModSearch({packId: () => 1, mcVersion: () => "1.20.1", mode, query, curseforgeAvailable })
  )!
  return {mode, query, search, scope}
}

describe("useModSearch", () => {
  beforeEach(() => {
    vi.useFakeTimers()
    modrinth.mockReset()
    curseforge.mockReset()
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it("does not overwrite newer results with a stale response", async () => {
    const first = deferred<ModSearchResponse>()
    modrinth.mockReturnValueOnce(first.promise)
    modrinth.mockResolvedValueOnce(response("new"))
    const {query, search} = setup()

    query.value = "old"
    await nextTick()
    await vi.advanceTimersByTimeAsync(SEARCH_DEBOUNCE_MS)
    expect(modrinth).toHaveBeenCalledTimes(1)

    query.value = "newer"
    await nextTick()
    await vi.advanceTimersByTimeAsync(SEARCH_DEBOUNCE_MS)
    expect(search.results.value.map((r) => r.slug)).toEqual(["new"])

    first.resolve(response("stale"))
    await vi.advanceTimersByTimeAsync(0)
    expect(search.results.value.map((r) => r.slug)).toEqual(["new"])
    expect(search.loading.value).toBe(false)
  })

  it("resets results when the mode switches", async () => {
    modrinth.mockResolvedValue(response("a"))
    const {mode, query, search} = setup()

    query.value = "sodium"
    await nextTick()
    await vi.advanceTimersByTimeAsync(SEARCH_DEBOUNCE_MS)
    expect(search.results.value).toHaveLength(1)

    mode.value = "url"
    await nextTick()
    expect(search.results.value).toEqual([])
    expect(search.loading.value).toBe(false)
    expect(search.hasSearched.value).toBe(false)
  })

  it("discards an in-flight response after switching to url mode", async () => {
    const pending = deferred<ModSearchResponse>()
    modrinth.mockReturnValueOnce(pending.promise)
    const {mode, query, search} = setup()

    query.value = "sodium"
    await nextTick()
    await vi.advanceTimersByTimeAsync(SEARCH_DEBOUNCE_MS)
    mode.value = "url"
    await nextTick()
    pending.resolve(response("late"))
    await vi.advanceTimersByTimeAsync(0)
    expect(search.results.value).toEqual([])
  })

  it("does not fire a request for whitespace or short queries", async () => {
    const {query, search} = setup()

    for (const q of ["   ", "a", " a "]) {
      query.value = q
      await nextTick()
      await vi.advanceTimersByTimeAsync(SEARCH_DEBOUNCE_MS * 2)
    }

    expect(modrinth).not.toHaveBeenCalled()
    expect(search.loading.value).toBe(false)
    expect(search.emptyState.value).toBe("short-query")
  })

  it("does not search curseforge when unavailable", async () => {
    const {query} = setup("curseforge", false)

    query.value = "jei"
    await nextTick()
    await vi.advanceTimersByTimeAsync(SEARCH_DEBOUNCE_MS)

    expect(curseforge).not.toHaveBeenCalled()
  })
})
