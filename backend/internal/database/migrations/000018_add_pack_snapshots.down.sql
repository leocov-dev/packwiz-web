DELETE
FROM permissions
WHERE name IN ('pack.snapshot.view', 'pack.snapshot.revert');

ALTER TABLE packs
    DROP COLUMN head_snapshot_id;

DROP TABLE pack_snapshots;
