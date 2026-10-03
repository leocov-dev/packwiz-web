-- Pack history: one full content snapshot per change to a non-draft pack.
-- Live snapshots form a single chain ending at packs.head_snapshot_id. A revert
-- marks the live snapshots after its target as abandoned; they are never deleted.

CREATE TABLE pack_snapshots
(
    id                     SERIAL PRIMARY KEY,
    pack_id                INTEGER     NOT NULL REFERENCES packs (id) ON DELETE CASCADE,
    seq                    INTEGER     NOT NULL,
    parent_id              INTEGER REFERENCES pack_snapshots (id),
    reason                 VARCHAR(32) NOT NULL CHECK (reason IN (
                                                                  'baseline', 'publish', 'pack_edit',
                                                                  'mod_add', 'mod_remove', 'mod_update',
                                                                  'mod_side', 'mod_option', 'mod_pin',
                                                                  'rehash', 'update_all', 'migrate',
                                                                  'migrate_mods')),
    detail                 JSONB       NOT NULL DEFAULT '{}',
    summary                JSONB       NOT NULL DEFAULT '{}',
    created_by             INTEGER     NOT NULL REFERENCES users (id),
    created_at             TIMESTAMPTZ NOT NULL DEFAULT now(),
    abandoned_at           TIMESTAMPTZ,
    abandoned_by           INTEGER REFERENCES users (id),
    abandoned_by_revert_to INTEGER REFERENCES pack_snapshots (id),
    schema_version         SMALLINT    NOT NULL,
    payload                JSONB       NOT NULL,
    payload_hash           CHAR(64)    NOT NULL,
    CONSTRAINT pack_snapshots_pack_seq_key UNIQUE (pack_id, seq)
);

CREATE INDEX idx_pack_snapshots_pack_live ON pack_snapshots (pack_id, seq DESC) WHERE abandoned_at IS NULL;
CREATE INDEX idx_pack_snapshots_parent_id ON pack_snapshots (parent_id);

ALTER TABLE packs
    ADD COLUMN head_snapshot_id INTEGER REFERENCES pack_snapshots (id);

INSERT INTO permissions (name, resource, description)
VALUES ('pack.snapshot.view', 'pack', 'View pack history and clone from a snapshot'),
       ('pack.snapshot.revert', 'pack', 'Revert the pack to a snapshot');

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
         JOIN permissions p ON (
    (r.name = 'system_admin' AND p.name IN ('pack.snapshot.view', 'pack.snapshot.revert'))
        OR (r.name = 'owner' AND p.name IN ('pack.snapshot.view', 'pack.snapshot.revert'))
        OR (r.name = 'editor' AND p.name IN ('pack.snapshot.view'))
    );
