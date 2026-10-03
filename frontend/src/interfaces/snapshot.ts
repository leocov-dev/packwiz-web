import {Type} from "class-transformer";

export type SnapshotAgainst = "parent" | "current"

export class SnapshotSummary {
  added!: number;
  removed!: number;
  changed!: number;
  packFields!: string[];
}

export class PackSnapshot {
  id!: number;
  seq!: number;
  parentId!: number | null;
  reason!: string;
  detail!: Record<string, unknown>;
  @Type(() => SnapshotSummary)
  summary!: SnapshotSummary;
  createdAt!: string;
  createdBy!: number;
  createdByUsername!: string;
  abandonedAt!: string | null;
  abandonedByRevertTo!: number | null;
  /** The snapshot the pack's history currently ends at. */
  isHead!: boolean;

  get isAbandoned(): boolean {
    return this.abandonedAt !== null && this.abandonedAt !== undefined;
  }
}

export class PackSnapshotListResponse {
  @Type(() => PackSnapshot)
  snapshots!: PackSnapshot[];
  total!: number;
  headId!: number | null;
  packStatus!: string;
}

export class SnapshotFieldChange {
  field!: string;
  from!: unknown;
  to!: unknown;
}

export class SnapshotModRef {
  slug!: string;
  name!: string;
  version!: string;
  fileName!: string;
}

export class SnapshotModChange extends SnapshotModRef {
  @Type(() => SnapshotFieldChange)
  changes!: SnapshotFieldChange[];
}

export class SnapshotDiff {
  @Type(() => SnapshotFieldChange)
  pack!: SnapshotFieldChange[];
  @Type(() => SnapshotModRef)
  added!: SnapshotModRef[];
  @Type(() => SnapshotModRef)
  removed!: SnapshotModRef[];
  @Type(() => SnapshotModChange)
  changed!: SnapshotModChange[];
}

export class PackSnapshotDetailResponse {
  @Type(() => PackSnapshot)
  snapshot!: PackSnapshot;
  against!: SnapshotAgainst;
  @Type(() => SnapshotDiff)
  diff!: SnapshotDiff;
}

export class RevertSnapshotResponse {
  changed!: boolean;
  headId!: number;
}

export class PruneSnapshotsResponse {
  deleted!: number;
}

export class RebaseSnapshotResponse {
  deleted!: number;
}
