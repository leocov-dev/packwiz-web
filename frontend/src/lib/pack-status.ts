import type {PackResponse} from "@/interfaces/pack.ts";

export type PackStatusKey = "draft" | "published" | "public" | "archived" | "warning"

/** The one status icon a pack shows: archived > draft > public > published. */
export const packStatusKey = (
  pack: Pick<PackResponse, "status" | "isArchived" | "isPublic">,
): PackStatusKey => {
  if (pack.isArchived) return "archived"
  if (pack.status === "draft") return "draft"
  return pack.isPublic ? "public" : "published"
}
