-- Per-pack outgoing webhook, called after the pack changes. Headers (which may
-- carry credentials) are stored encrypted as a JSON object.
CREATE TABLE pack_webhooks
(
    pack_id          INTEGER PRIMARY KEY REFERENCES packs (id) ON DELETE CASCADE,
    enabled          BOOLEAN     NOT NULL DEFAULT FALSE,
    url              TEXT        NOT NULL DEFAULT '',
    method           TEXT        NOT NULL DEFAULT 'POST',
    headers_enc      TEXT        NOT NULL DEFAULT '',
    last_status_code INTEGER     NOT NULL DEFAULT 0,
    last_error       TEXT        NOT NULL DEFAULT '',
    last_attempt_at  TIMESTAMPTZ,
    last_success_at  TIMESTAMPTZ,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO permissions (name, resource, description)
VALUES ('pack.webhook.manage', 'pack', 'Configure the webhook called when the pack changes');

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
         JOIN permissions p ON (
    (r.name = 'system_admin' AND p.name IN ('pack.webhook.manage'))
        OR (r.name = 'owner' AND p.name IN ('pack.webhook.manage'))
    );
