-- One row per request for a pack's MultiMC / Prism instance zip, successful or
-- not. Same shape as pack_accesses, but kept apart so pack sync metrics only
-- count pack.toml. Shown on the system admin page only.
-- Pruned by the same AUDIT_RETENTION_DAYS setting as audits.

CREATE TABLE instance_downloads
(
    id          SERIAL PRIMARY KEY,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    pack_id     INTEGER REFERENCES packs (id) ON DELETE CASCADE,
    pack_slug   VARCHAR(255) NOT NULL,
    user_id     INTEGER REFERENCES users (id) ON DELETE SET NULL,
    ip_address  VARCHAR(64)  NOT NULL DEFAULT '',
    user_agent  VARCHAR(512) NOT NULL DEFAULT '',
    status_code INTEGER      NOT NULL,
    success     BOOLEAN      NOT NULL
);

CREATE INDEX idx_instance_downloads_created_at ON instance_downloads (created_at);
CREATE INDEX idx_instance_downloads_pack_created ON instance_downloads (pack_id, created_at);
