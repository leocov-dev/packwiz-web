
-- Create users table
CREATE TABLE users (
                       id SERIAL PRIMARY KEY,
                       created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                       updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                       deleted_at TIMESTAMP,
                       username VARCHAR(255) UNIQUE NOT NULL,
                       full_name VARCHAR(255),
                       email VARCHAR(255) UNIQUE,
                       password VARCHAR(255),
                       is_admin BOOLEAN NOT NULL DEFAULT FALSE,
                       identity_provider VARCHAR(255),
                       link_token VARCHAR(255) UNIQUE,
                       session_key VARCHAR(255)
);

-- Create index for soft deletes
CREATE INDEX idx_users_deleted_at ON users(deleted_at);

-- Create packs table
CREATE TABLE packs (
                       slug VARCHAR(255) PRIMARY KEY,
                       name VARCHAR(255) NOT NULL,
                       created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                       updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                       deleted_at TIMESTAMP,
                       description TEXT,
                       created_by INTEGER NOT NULL,
                       updated_by INTEGER NOT NULL,
                       is_public BOOLEAN NOT NULL DEFAULT FALSE,
                       status VARCHAR(50) NOT NULL DEFAULT 'draft',
                       mc_version VARCHAR(50),
                       loader VARCHAR(50),
                       loader_version VARCHAR(50),
                       acceptable_game_versions JSON,
                       version VARCHAR(50),
                       pack_format VARCHAR(50),
                       FOREIGN KEY (created_by) REFERENCES users(id),
                       FOREIGN KEY (updated_by) REFERENCES users(id)
);

-- Create index for soft deletes
CREATE INDEX idx_packs_deleted_at ON packs(deleted_at);

-- Create mods table
CREATE TABLE mods (
                      id SERIAL PRIMARY KEY,
                      slug VARCHAR(255) NOT NULL,
                      pack_slug VARCHAR(255) NOT NULL,
                      name VARCHAR(255) NOT NULL,
                      file_name VARCHAR(255),
                      side VARCHAR(50),
                      pinned BOOLEAN NOT NULL DEFAULT FALSE,
                      download JSON,
                      hash_format VARCHAR(50) DEFAULT 'sha256',
                      alias VARCHAR(255),
                      type VARCHAR(50) DEFAULT 'mods',
                      source VARCHAR(255),
                      update JSON,
                      preserve BOOLEAN DEFAULT FALSE,
                      created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                      updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                      created_by INTEGER NOT NULL,
                      updated_by INTEGER NOT NULL,
                      FOREIGN KEY (pack_slug) REFERENCES packs(slug),
                      FOREIGN KEY (created_by) REFERENCES users(id),
                      FOREIGN KEY (updated_by) REFERENCES users(id)
);

-- Create unique index for pack_slug + mod_slug combination
CREATE UNIQUE INDEX idx_pack_mod_slug ON mods(pack_slug, slug);

-- Create pack_users table (junction table for pack permissions)
CREATE TABLE pack_users (
                            pack_slug VARCHAR(255) NOT NULL,
                            user_id INTEGER NOT NULL,
                            created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                            permission VARCHAR(50) NOT NULL,
                            PRIMARY KEY (pack_slug, user_id),
                            FOREIGN KEY (pack_slug) REFERENCES packs(slug),
                            FOREIGN KEY (user_id) REFERENCES users(id)
);

-- Create unique index for pack_slug + user_id combination
CREATE UNIQUE INDEX idx_pack_slug_user_id ON pack_users(pack_slug, user_id);

-- Create audits table
CREATE TABLE audits (
                        id SERIAL PRIMARY KEY,
                        created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                        user_id INTEGER NOT NULL,
                        action VARCHAR(255) NOT NULL,
                        action_params JSON,
                        ip_address VARCHAR(45),
                        FOREIGN KEY (user_id) REFERENCES users(id)
);