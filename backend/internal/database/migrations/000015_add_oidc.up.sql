-- OIDC identity providers, managed by admins. The client secret is stored
-- AES-GCM encrypted (base64) and is never returned by the API.
CREATE TABLE oidc_providers
(
    id                SERIAL PRIMARY KEY,
    slug              VARCHAR(64)  NOT NULL UNIQUE,
    display_name      VARCHAR(255) NOT NULL,
    issuer_url        VARCHAR(512) NOT NULL,
    client_id         VARCHAR(512) NOT NULL,
    client_secret_enc TEXT         NOT NULL DEFAULT '',
    scopes            VARCHAR(512) NOT NULL DEFAULT 'openid profile email',
    enabled           BOOLEAN      NOT NULL DEFAULT TRUE,
    auto_create_users BOOLEAN      NOT NULL DEFAULT FALSE,
    link_by_email     BOOLEAN      NOT NULL DEFAULT FALSE,
    created_at        TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- A user may have several external identities, matched on (provider, subject).
CREATE TABLE user_identities
(
    id            SERIAL PRIMARY KEY,
    user_id       INTEGER      NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    provider_id   INTEGER      NOT NULL REFERENCES oidc_providers (id) ON DELETE CASCADE,
    subject       VARCHAR(512) NOT NULL,
    email         VARCHAR(320) NOT NULL DEFAULT '',
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    last_login_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX idx_user_identities_provider_subject ON user_identities (provider_id, subject);
-- one identity per user per provider, enforced here to close check-then-insert races
CREATE UNIQUE INDEX idx_user_identities_user_provider ON user_identities (user_id, provider_id);
