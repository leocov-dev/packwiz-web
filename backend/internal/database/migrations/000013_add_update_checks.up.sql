-- Per-mod results of the on-demand "check for updates" job. Replaced wholesale
-- (delete + insert) per pack each time a check finishes.
CREATE TABLE mod_update_checks
(
    id               SERIAL PRIMARY KEY,
    pack_id          INTEGER     NOT NULL REFERENCES packs (id) ON DELETE CASCADE,
    mod_id           INTEGER     NOT NULL REFERENCES mods (id) ON DELETE CASCADE,
    update_available BOOLEAN     NOT NULL DEFAULT false,
    update_string    TEXT        NOT NULL DEFAULT '',
    error            TEXT        NOT NULL DEFAULT '',
    checked_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT mod_update_checks_pack_mod_key UNIQUE (pack_id, mod_id)
);

-- Latest check job per pack (at most one row per pack).
CREATE TABLE pack_update_check_runs
(
    pack_id     INTEGER PRIMARY KEY REFERENCES packs (id) ON DELETE CASCADE,
    job_id      BIGINT      NOT NULL DEFAULT 0,
    status      TEXT        NOT NULL,
    started_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    finished_at TIMESTAMPTZ,
    error       TEXT        NOT NULL DEFAULT ''
);
