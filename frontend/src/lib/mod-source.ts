import type {AddModRequest} from "@/interfaces/requests.ts";

export type ModSource = "Curseforge" | "Modrinth" | "Github" | ""

const SOURCE_HOSTS: Array<[string, ModSource]> = [
  ["curseforge.com", "Curseforge"],
  ["modrinth.com", "Modrinth"],
  ["github.com", "Github"],
]

const HAS_SCHEME = /^[a-z][a-z0-9+.-]*:\/\//i

/**
 * Normalises a pasted URL: trims, and prepends https:// when no scheme is present
 * (e.g. "modrinth.com/mod/foo"). Returns the input trimmed otherwise.
 */
export function normalizeUrl(url: string): string {
  const trimmed = (url ?? "").trim()
  if (!trimmed || HAS_SCHEME.test(trimmed)) {
    return trimmed
  }
  return `https://${trimmed}`
}

export function parseUrl(url: string): ModSource {
  const normalized = normalizeUrl(url)
  if (!normalized) {
    return ""
  }

  let parsed: URL
  try {
    parsed = new URL(normalized)
  } catch {
    return ""
  }

  if (parsed.protocol !== "http:" && parsed.protocol !== "https:") {
    return ""
  }

  const host = parsed.hostname.toLowerCase()
  for (const [domain, source] of SOURCE_HOSTS) {
    if (host === domain || host.endsWith("." + domain)) {
      return source
    }
  }

  return ""
}

const CURSEFORGE_FILTER_LOADERS = ["fabric", "forge", "neoforge"]

/**
 * Caption describing the filters the backend really applies for a search source.
 * CurseForge only filters by loader for fabric/forge/neoforge; other loaders
 * (e.g. quilt) are searched without a loader filter, so they are omitted.
 */
export function searchFilterCaption(
  source: "modrinth" | "curseforge",
  mcVersion: string | undefined,
  loader: string | undefined,
): string {
  const parts: string[] = []
  if (mcVersion) {
    parts.push(`Minecraft ${mcVersion}`)
  }
  const loaderName = (loader ?? "").trim()
  if (loaderName) {
    const applies = source === "modrinth" || CURSEFORGE_FILTER_LOADERS.includes(loaderName.toLowerCase())
    if (applies) {
      parts.push(loaderName.charAt(0).toUpperCase() + loaderName.slice(1).toLowerCase())
    }
  }
  return parts.length > 0 ? `Filtering for ${parts.join(" \u00b7 ")}` : ""
}

export type BuildRequestResult =
  | { request: AddModRequest }
  | { error: string }

export function buildRequest(modSource: ModSource, modUrl: string): BuildRequestResult {
  if (modSource === "Curseforge") {
    return {
      request: {
        curseforge: {
          url: modUrl,
        }
      }
    }
  } else if (modSource === "Modrinth") {
    return {
      request: {
        modrinth: {
          url: modUrl,
        }
      }
    }
  } else if (modSource === "Github") {
    return {
      request: {
        github: {
          url: modUrl,
        }
      }
    }
  }

  return {error: `Invalid mod source: ${modSource}`}
}

function asRecord(value: unknown): Record<string, unknown> | undefined {
  return typeof value === "object" && value !== null ? value as Record<string, unknown> : undefined
}

export function isSearchResultInstalled(
  result: { slug?: string; projectId?: string; installed?: boolean },
  installedMods?: Array<{ slug?: string; update?: Record<string, unknown> }>,
): boolean {
  if (result.installed) {
    return true
  }
  if (!installedMods || installedMods.length === 0) {
    return false
  }
  const resultSlug = result.slug?.toLowerCase()
  const resultId = result.projectId ? String(result.projectId) : undefined
  return installedMods.some((mod) => {
    if (resultSlug && mod.slug && mod.slug.toLowerCase() === resultSlug) {
      return true
    }
    const update = mod.update
    const modrinthUpdate = asRecord(update?.modrinth)
    const curseforgeUpdate = asRecord(update?.curseforge)
    const updateModId = update?.['mod-id'] || modrinthUpdate?.['mod-id']
    const updateProjectId = update?.['project-id'] || curseforgeUpdate?.['project-id']

    if (resultId && updateModId && String(updateModId) === resultId) {
      return true
    }
    if (resultId && updateProjectId && String(updateProjectId) === resultId) {
      return true
    }
    return false
  })
}

export function modPageUrl(source: "modrinth" | "curseforge", slug: string): string {
  if (source === "curseforge") {
    return `https://www.curseforge.com/minecraft/mc-mods/${slug}`
  }
  return `https://modrinth.com/mod/${slug}`
}

export function addedKey(source: "modrinth" | "curseforge", slug: string): string {
  return `${source}:${slug}`
}

export function normalizeSearchQuery(query: string | null | undefined): string {
  return (query ?? "").trim()
}

export type ResultState = "available" | "installed" | "added"

export function getResultState(
  result: { slug?: string; projectId?: string; installed?: boolean },
  installedMods: Array<{ slug?: string; update?: Record<string, unknown> }> | undefined,
  addedKeys: readonly string[],
  source: "modrinth" | "curseforge",
): ResultState {
  if (result.slug && addedKeys.includes(addedKey(source, result.slug))) {
    return "added"
  }
  return isSearchResultInstalled(result, installedMods) ? "installed" : "available"
}

export type SearchEmptyState = "short-query" | "no-results" | "none"

export function searchEmptyState(query: string, loading: boolean, resultCount: number, hasSearched: boolean): SearchEmptyState {
  if (loading || resultCount > 0) {
    return "none"
  }
  const trimmed = normalizeSearchQuery(query)
  if (trimmed.length < 2) {
    return "short-query"
  }
  return hasSearched ? "no-results" : "none"
}
