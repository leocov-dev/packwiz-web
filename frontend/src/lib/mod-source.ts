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
