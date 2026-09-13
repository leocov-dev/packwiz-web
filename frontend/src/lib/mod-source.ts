import type {AddModRequest} from "@/interfaces/requests.ts";

export type ModSource = "Curseforge" | "Modrinth" | "Github" | ""

export function parseUrl(url: string): ModSource {
  if (!url) {
    return ""
  }
  if (url.includes("curseforge.com")) {
    return "Curseforge"
  } else if (url.includes("modrinth.com")) {
    return "Modrinth"
  } else if (url.includes("github.com")) {
    return "Github"
  } else {
    return ""
  }
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

export function isSearchResultInstalled(
  result: { slug?: string; projectId?: string; installed?: boolean },
  installedMods?: Array<{ slug?: string; update?: Record<string, any> }>,
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
    const update = mod.update as any
    const modrinthUpdate = update?.modrinth
    const curseforgeUpdate = update?.curseforge
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
