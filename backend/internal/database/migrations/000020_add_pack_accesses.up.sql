-- One row per request for a pack's pack.toml, successful or not, for access
-- metrics. pack_id is null when the requested slug matched no pack. user_id is
-- only set when a private-pack link token resolved to a user.
-- Pruned by the same AUDIT_RETENTION_DAYS setting as audits.

CREATE TABLE pack_accesses
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

CREATE INDEX idx_pack_accesses_created_at ON pack_accesses (created_at);
CREATE INDEX idx_pack_accesses_pack_created ON pack_accesses (pack_id, created_at);
