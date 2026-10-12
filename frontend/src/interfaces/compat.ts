import {Type} from "class-transformer";

export type CompatStatus =
  | "unknown" | "queued" | "scanning" | "stale" | "ok" | "warnings" | "errors" | "failed"

export type CompatSeverity = "error" | "warning" | "info"

/** Counts of findings per severity. */
export class CompatCounts {
  errors!: number;
  warnings!: number;
  info!: number;
}

export class CompatFindingItem {
  /** Mod the finding is about; null when it is not tied to an installed mod. */
  modId!: number | null;
  slug!: string;
  severity!: CompatSeverity;
  code!: string;
  message!: string;
  /** Free-form key/value context (dependency id, required range, found version, ...). */
  detail: Record<string, string> = {};
}

export class CompatResponse {
  status!: CompatStatus;
  @Type(() => CompatCounts)
  counts!: CompatCounts;
  /** Mods that could not be checked. */
  unscanned!: number;
  scannedAt!: string | null;
  /** "" | "scan failed" | "scan timed out" */
  runError!: string;
  @Type(() => CompatFindingItem)
  findings: CompatFindingItem[] = [];
}

export class CompatScanResponse {
  status!: CompatStatus;
  enqueued!: boolean;
}

/** Compact result carried on each pack of the pack list and the pack detail. */
export class CompatSummary {
  status!: CompatStatus;
  errors!: number;
  warnings!: number;
}
