export type ChangelistPeriod = "day" | "month" | "year" | "pending"

export interface ChangelistFieldChange {
  field: string
  from: unknown
  to: unknown
}

export interface ChangelistMod {
  slug: string
  name: string
  version: string
}

export interface ChangelistModChange {
  slug: string
  name: string
  fromVersion: string
  toVersion: string
  updated: boolean
  fields: string[]
}

/** The net change of a pack over one period (UTC dates, YYYY-MM-DD). */
export interface ChangelistEntry {
  period: ChangelistPeriod
  start: string
  end: string
  /** No earlier history: describes the pack as first recorded. */
  initial: boolean
  modCount: number
  pack: ChangelistFieldChange[]
  added: ChangelistMod[]
  removed: ChangelistMod[]
  changed: ChangelistModChange[]
}

export interface PackChangelist {
  /** Newest first. */
  entries: ChangelistEntry[]
}
