CREATE TABLE mod_migration_results
(
    id               SERIAL PRIMARY KEY,
    job_id           BIGINT       NOT NULL,
    pack_id          INTEGER      NOT NULL REFERENCES packs (id),
    mod_id           INTEGER      NOT NULL,
    slug             TEXT         NOT NULL,
    name             TEXT         NOT NULL,
    pinned           BOOLEAN      NOT NULL DEFAULT false,
    update_available BOOLEAN      NOT NULL DEFAULT false,
    update_string    TEXT,
    incompatible     BOOLEAN      NOT NULL DEFAULT false,
    error            TEXT,
    created_at       TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX idx_mod_migration_results_job_id ON mod_migration_results (job_id);
