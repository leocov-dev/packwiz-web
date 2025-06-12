
-- Create users table
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    username VARCHAR(255) UNIQUE NOT NULL,
    full_name VARCHAR(255),
    email VARCHAR(320) UNIQUE,
    password VARCHAR(255),
    is_admin BOOLEAN NOT NULL DEFAULT FALSE,
    link_token VARCHAR(255) UNIQUE,
    session_key VARCHAR(255)
);

CREATE INDEX idx_users_deleted_at ON users(deleted_at);
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_session_key ON users(session_key);

-- Create packs table
CREATE TABLE packs (
    slug VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    description TEXT,
    created_by INTEGER NOT NULL,
    updated_by INTEGER NOT NULL,
    is_public BOOLEAN NOT NULL DEFAULT FALSE,
    status VARCHAR(50) NOT NULL DEFAULT 'draft',
    mc_version VARCHAR(50),
    loader VARCHAR(50),
    loader_version VARCHAR(50),
    acceptable_game_versions JSONB,
    version VARCHAR(50),
    pack_format VARCHAR(50),
    FOREIGN KEY (created_by) REFERENCES users(id),
    FOREIGN KEY (updated_by) REFERENCES users(id)
);

CREATE INDEX idx_packs_slug ON packs(slug);
CREATE INDEX idx_packs_deleted_at ON packs(deleted_at);
CREATE INDEX idx_packs_is_public ON packs(is_public);
CREATE INDEX idx_packs_status ON packs(status);

-- Create mods table
CREATE TABLE mods (
    id SERIAL PRIMARY KEY,
    slug VARCHAR(255) NOT NULL,
    pack_slug VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    file_name VARCHAR(255),
    side VARCHAR(50),
    pinned BOOLEAN NOT NULL DEFAULT FALSE,
    download JSONB,
    hash_format VARCHAR(50) DEFAULT 'sha256',
    alias VARCHAR(255),
    type VARCHAR(50) DEFAULT 'mods',
    source VARCHAR(255),
    update JSONB,
    preserve BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by INTEGER NOT NULL,
    updated_by INTEGER NOT NULL,
    FOREIGN KEY (pack_slug) REFERENCES packs(slug),
    FOREIGN KEY (created_by) REFERENCES users(id),
    FOREIGN KEY (updated_by) REFERENCES users(id)
);

CREATE UNIQUE INDEX idx_mods_pack_mod_slug ON mods(pack_slug, slug);

-- Create pack_users table (junction table for pack permissions)
CREATE TABLE pack_users (
    pack_slug VARCHAR(255) NOT NULL,
    user_id INTEGER NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    permission SMALLINT NOT NULL DEFAULT 1 CHECK ( permission BETWEEN 0 AND 999 ),
    PRIMARY KEY (pack_slug, user_id),
    FOREIGN KEY (pack_slug) REFERENCES packs(slug),
    FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE UNIQUE INDEX idx_pack_users_pack_slug_user_id ON pack_users(pack_slug, user_id);
CREATE UNIQUE INDEX idx_pack_users_user_id_pack_slug ON pack_users(user_id, pack_slug);

-- Create audits table
CREATE TABLE audits (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    user_id INTEGER NOT NULL,
    action VARCHAR(255) NOT NULL,
    action_params JSONB,
    ip_address VARCHAR(45)
);

CREATE INDEX idx_audit_user_id ON audits(user_id);
CREATE INDEX idx_audit_action ON audits(action);
CREATE INDEX idx_audit_ip_address ON audits(ip_address);
CREATE INDEX idx_audits_created_at ON audits(created_at);
