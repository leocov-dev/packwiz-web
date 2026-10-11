-- One "add mods from a list" request. The BulkAddArgs job resolves its items.
CREATE TABLE pack_bulk_add_runs
(
    id          SERIAL PRIMARY KEY,
    pack_id     INTEGER     NOT NULL REFERENCES packs (id) ON DELETE CASCADE,
    job_id      BIGINT      NOT NULL DEFAULT 0,
    status      TEXT        NOT NULL,
    created_by  INTEGER     NOT NULL REFERENCES users (id),
    error       TEXT        NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    started_at  TIMESTAMPTZ,
    finished_at TIMESTAMPTZ
);

CREATE INDEX idx_pack_bulk_add_runs_pack_id ON pack_bulk_add_runs (pack_id, id DESC);

-- One line of a bulk add request, in list order (position is 1-based).
CREATE TABLE pack_bulk_add_items
(
    id                      SERIAL PRIMARY KEY,
    run_id                  INTEGER     NOT NULL REFERENCES pack_bulk_add_runs (id) ON DELETE CASCADE,
    position                INTEGER     NOT NULL,
    line                    TEXT        NOT NULL,
    status                  TEXT        NOT NULL,
    source                  TEXT        NOT NULL DEFAULT '',
    slug                    TEXT        NOT NULL DEFAULT '',
    name                    TEXT        NOT NULL DEFAULT '',
    project_id              TEXT        NOT NULL DEFAULT '',
    reason                  TEXT        NOT NULL DEFAULT '',
    conflicts_with_position INTEGER     NOT NULL DEFAULT 0,
    candidates              JSONB       NOT NULL DEFAULT '[]',
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT pack_bulk_add_items_run_position_key UNIQUE (run_id, position)
);

-- Allow the 'mod_bulk_add' snapshot reason (mods added from a list).
ALTER TABLE pack_snapshots DROP CONSTRAINT pack_snapshots_reason_check;
ALTER TABLE pack_snapshots ADD CONSTRAINT pack_snapshots_reason_check CHECK (reason IN (
    'baseline', 'publish', 'pack_edit',
    'mod_add', 'mod_bulk_add', 'mod_remove', 'mod_update',
    'mod_side', 'mod_option', 'mod_pin', 'mod_uplift',
    'rehash', 'update_all', 'migrate',
    'migrate_mods'));
