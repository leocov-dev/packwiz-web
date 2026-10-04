import {compareMinecraftVersions} from "@/lib/mc-version.ts"

export interface TargetPack {
  status: string
  mcVersion: string
  loader: string
}

export interface TargetCheck {
  /** Why the change is refused (mirrors the server), or null. */
  blocked: string | null
  /** The change reaches clients' instances and needs confirmation. */
  confirm: boolean
}

/**
 * Mirrors the server rule for published packs: a loader change or an older
 * Minecraft version is refused (clone instead); any other Minecraft change is
 * allowed after confirmation. Drafts aren't served, so anything goes.
 * `mcVersion` may be a "latest" sentinel, which can't be compared and is
 * treated as a change.
 */
export function checkTargetChange(pack: TargetPack, mcVersion: string, loader: string): TargetCheck {
  if (pack.status !== "published") return {blocked: null, confirm: false}

  if (loader && loader.toLowerCase() !== pack.loader.toLowerCase()) {
    return {
      blocked: `A published pack can't change its loader (${pack.loader} to ${loader}). Clone it to a new pack instead, or convert it to a draft first.`,
      confirm: false,
    }
  }
  if (compareMinecraftVersions(mcVersion, pack.mcVersion) === -1) {
    return {
      blocked: `A published pack can't move to an older Minecraft version (${pack.mcVersion} to ${mcVersion}). Clone it to a new pack instead, or convert it to a draft first.`,
      confirm: false,
    }
  }
  return {blocked: null, confirm: !!mcVersion && mcVersion !== pack.mcVersion}
}
