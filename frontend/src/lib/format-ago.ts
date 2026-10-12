/**
 * "<prefix> just now" / "<prefix> 5 min ago" / "... 3 h ago" / "... 2 d ago".
 * Empty string when the timestamp is missing or unparseable.
 */
export function formatAgo(prefix: string, timestamp: string | null | undefined, now: number = Date.now()): string {
  const t = Date.parse(timestamp ?? "")
  if (Number.isNaN(t)) return ""
  const minutes = Math.floor(Math.max(0, now - t) / 60000)
  if (minutes < 1) return `${prefix} just now`
  if (minutes < 60) return `${prefix} ${minutes} min ago`
  const hours = Math.floor(minutes / 60)
  if (hours < 24) return `${prefix} ${hours} h ago`
  return `${prefix} ${Math.floor(hours / 24)} d ago`
}
