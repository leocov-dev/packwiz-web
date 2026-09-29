const compact = new Intl.NumberFormat("en", {notation: "compact", maximumFractionDigits: 1})

/** "1.2M downloads"; empty string when the count is missing or not positive. */
export function formatDownloads(count?: number): string {
  if (count === undefined || !Number.isFinite(count) || count <= 0) {
    return ""
  }
  return `${compact.format(count)} ${count === 1 ? "download" : "downloads"}`
}

/** "by alice · 1.2M downloads"; missing parts are omitted. */
export function formatAuthorLine(author?: string, downloads?: number): string {
  const parts = [author ? `by ${author}` : "", formatDownloads(downloads)]
  return parts.filter(Boolean).join(" · ")
}

/** Case-insensitive check that a loader chip matches the pack's loader. */
export function isPackLoader(loader: string, packLoader?: string): boolean {
  return !!packLoader && loader.toLowerCase() === packLoader.toLowerCase()
}

/** Split a list into the first `max` items and the count of the rest. */
export function limitChips<T>(items: T[] | undefined, max: number): { shown: T[], hidden: number } {
  const list = items ?? []
  return {shown: list.slice(0, max), hidden: Math.max(0, list.length - max)}
}

/** "game-mechanics" -> "Game Mechanics". Only for slug-style names (Modrinth). */
export function prettyCategory(slug: string): string {
  return slug
    .split(/[-_\s]+/)
    .filter(Boolean)
    .map(w => w.charAt(0).toUpperCase() + w.slice(1))
    .join(" ")
}
