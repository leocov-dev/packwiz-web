-- Informational mod jar compatibility check. Jars are never retained: the
-- parsed loader metadata is the cache, shared across packs by jar hash.

-- Parsed jar metadata keyed by the jar hash. Parse failures are cached too
-- (meta NULL, error set) so an unreadable jar is not downloaded again on every
-- scan. Negative-result retry policy: a row with an error is re-fetched when
-- its parser_version is outdated or when parsed_at is older than the retry
-- window chosen by the scanner; a successful row never expires except by prune.
CREATE TABLE mod_jar_meta
(
    hash_format    TEXT        NOT NULL,
    hash           TEXT        NOT NULL,
    parser_version INTEGER     NOT NULL,
    meta           JSONB,
    loaders        TEXT[]      NOT NULL DEFAULT '{}',
    error          TEXT        NOT NULL DEFAULT '',
    parsed_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_used_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (hash_format, hash)
);

CREATE INDEX idx_mod_jar_meta_last_used_at ON mod_jar_meta (last_used_at);

-- Latest compatibility scan per pack (at most one row per pack).
CREATE TABLE pack_compat_runs
(
    pack_id         INTEGER PRIMARY KEY REFERENCES packs (id) ON DELETE CASCADE,
    job_id          BIGINT      NOT NULL DEFAULT 0,
    status          TEXT        NOT NULL,
    target_hash     TEXT        NOT NULL DEFAULT '',
    error_count     INTEGER     NOT NULL DEFAULT 0,
    warning_count   INTEGER     NOT NULL DEFAULT 0,
    info_count      INTEGER     NOT NULL DEFAULT 0,
    unscanned_count INTEGER     NOT NULL DEFAULT 0,
    started_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    finished_at     TIMESTAMPTZ,
    error           TEXT        NOT NULL DEFAULT ''
);

-- Findings of the latest completed scan. Replaced wholesale per pack, in one
-- transaction, when a run completes.
CREATE TABLE pack_compat_findings
(
    id       SERIAL PRIMARY KEY,
    pack_id  INTEGER NOT NULL REFERENCES packs (id) ON DELETE CASCADE,
    mod_id   INTEGER REFERENCES mods (id) ON DELETE CASCADE,
    slug     TEXT    NOT NULL DEFAULT '',
    severity TEXT    NOT NULL,
    code     TEXT    NOT NULL,
    message  TEXT    NOT NULL DEFAULT '',
    detail   JSONB
);

CREATE INDEX idx_pack_compat_findings_pack_id ON pack_compat_findings (pack_id);
