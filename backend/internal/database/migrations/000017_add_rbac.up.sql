-- RBAC: permission <- role_permissions <- role. See .plan/rbac.md.

CREATE TABLE roles
(
    id          SERIAL PRIMARY KEY,
    name        VARCHAR(64) UNIQUE NOT NULL,
    description TEXT               NOT NULL DEFAULT '',
    scope       VARCHAR(16)        NOT NULL CHECK (scope IN ('system', 'pack')),
    assignable  BOOLEAN            NOT NULL DEFAULT TRUE
);

CREATE TABLE permissions
(
    id          SERIAL PRIMARY KEY,
    name        VARCHAR(64) UNIQUE NOT NULL,
    description TEXT               NOT NULL DEFAULT '',
    resource    VARCHAR(16)        NOT NULL CHECK (resource IN ('pack', 'global'))
);

CREATE TABLE role_permissions
(
    role_id       INTEGER NOT NULL REFERENCES roles (id) ON DELETE CASCADE,
    permission_id INTEGER NOT NULL REFERENCES permissions (id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
);

CREATE INDEX idx_role_permissions_permission_id ON role_permissions (permission_id);

CREATE TABLE user_roles
(
    user_id INTEGER NOT NULL REFERENCES users (id),
    role_id INTEGER NOT NULL REFERENCES roles (id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, role_id)
);

CREATE INDEX idx_user_roles_role_id ON user_roles (role_id);

-- Pack roles may only hold pack-resource permissions.
CREATE FUNCTION check_role_permission_scope() RETURNS trigger AS
$$
BEGIN
    IF EXISTS (SELECT 1
               FROM roles r,
                    permissions p
               WHERE r.id = NEW.role_id
                 AND p.id = NEW.permission_id
                 AND r.scope = 'pack'
                 AND p.resource <> 'pack') THEN
        RAISE EXCEPTION 'pack roles may only contain pack permissions';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_role_permissions_scope
    BEFORE INSERT OR UPDATE
    ON role_permissions
    FOR EACH ROW
EXECUTE FUNCTION check_role_permission_scope();

-- user_roles holds system roles only.
CREATE FUNCTION check_user_role_scope() RETURNS trigger AS
$$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM roles WHERE id = NEW.role_id AND scope = 'system') THEN
        RAISE EXCEPTION 'user_roles may only reference system roles';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_user_roles_scope
    BEFORE INSERT OR UPDATE
    ON user_roles
    FOR EACH ROW
EXECUTE FUNCTION check_user_role_scope();

-- Seed permissions
INSERT INTO permissions (name, resource, description)
VALUES ('pack.create', 'global', 'Create packs'),
       ('user.view', 'global', 'View the admin user list and user details'),
       ('user.lookup', 'global', 'Search users for the collaborator picker'),
       ('user.create', 'global', 'Create users'),
       ('user.manage', 'global', 'Edit, deactivate and reactivate users'),
       ('user.roles.assign', 'global', 'Assign system roles to users'),
       ('audit.view', 'global', 'View audits'),
       ('oidc.manage', 'global', 'Manage OIDC providers'),
       ('pack.consume', 'pack', 'Access pack files through a personal link'),
       ('pack.view', 'pack', 'View the pack, its mods and update results'),
       ('pack.link', 'pack', 'Get the personal link'),
       ('pack.mod.add', 'pack', 'Add mods'),
       ('pack.mod.remove', 'pack', 'Remove mods'),
       ('pack.mod.update', 'pack', 'Update mods'),
       ('pack.mod.configure', 'pack', 'Change mod side, options and pinning'),
       ('pack.updates.check', 'pack', 'Trigger an update check'),
       ('pack.migrate', 'pack', 'Migrate the pack'),
       ('pack.rehash', 'pack', 'Rehash all mods'),
       ('pack.info.edit', 'pack', 'Edit pack name, description and version'),
       ('pack.publish', 'pack', 'Publish the pack or convert it to draft'),
       ('pack.visibility', 'pack', 'Make the pack public or private'),
       ('pack.archive', 'pack', 'Archive or unarchive the pack'),
       ('pack.users.view', 'pack', 'List collaborators'),
       ('pack.users.manage', 'pack', 'Add, remove and change collaborators');

-- Seed roles
INSERT INTO roles (name, scope, assignable, description)
VALUES ('system_admin', 'system', TRUE, 'Every global and pack permission'),
       ('project_admin', 'system', TRUE, 'Can create packs and look up users'),
       ('user', 'system', TRUE, 'Default role, no extra permissions'),
       ('viewer', 'pack', TRUE, 'Can view the pack and install from it'),
       ('editor', 'pack', TRUE, 'Can view and edit mods'),
       ('owner', 'pack', FALSE, 'Full control of the pack');

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
         JOIN permissions p ON (
    r.name = 'system_admin'
        OR (r.name = 'project_admin' AND p.name IN ('pack.create', 'user.lookup'))
        OR (r.name = 'viewer' AND p.name IN ('pack.consume', 'pack.view', 'pack.link'))
        OR (r.name = 'editor' AND p.name IN (
                                             'pack.consume', 'pack.view', 'pack.link',
                                             'pack.mod.add', 'pack.mod.remove', 'pack.mod.update',
                                             'pack.mod.configure', 'pack.updates.check', 'pack.users.view'))
        OR (r.name = 'owner' AND p.resource = 'pack')
    );

-- Users: superuser flag replaces is_admin
ALTER TABLE users
    ADD COLUMN is_superuser BOOLEAN NOT NULL DEFAULT FALSE;

-- The bootstrap admin is the superuser (see UpsertDefaultAdminUser).
UPDATE users
SET is_superuser = TRUE
WHERE username = 'admin';

CREATE UNIQUE INDEX idx_users_single_superuser ON users (is_superuser) WHERE is_superuser;

-- Other former admins become system_admin; everyone else (superuser excluded) gets 'user'.
INSERT INTO user_roles (user_id, role_id)
SELECT u.id, r.id
FROM users u
         JOIN roles r ON r.name = CASE WHEN u.is_admin THEN 'system_admin' ELSE 'user' END
WHERE NOT u.is_superuser;

ALTER TABLE users
    DROP COLUMN is_admin;

-- pack_users: permission level -> pack role.
-- Static (link-only) grants are dropped; the user account is untouched.
DELETE
FROM pack_users
WHERE permission = 1;

ALTER TABLE pack_users
    ADD COLUMN role_id INTEGER REFERENCES roles (id);

UPDATE pack_users pu
SET role_id = r.id
FROM roles r
WHERE r.name = CASE pu.permission WHEN 10 THEN 'viewer' WHEN 20 THEN 'editor' WHEN 30 THEN 'owner' END;

ALTER TABLE pack_users
    ALTER COLUMN role_id SET NOT NULL;

ALTER TABLE pack_users
    DROP COLUMN permission;

CREATE INDEX idx_pack_users_role_id ON pack_users (role_id);

-- pack_users holds pack roles only.
CREATE FUNCTION check_pack_user_role_scope() RETURNS trigger AS
$$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM roles WHERE id = NEW.role_id AND scope = 'pack') THEN
        RAISE EXCEPTION 'pack_users may only reference pack roles';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_pack_users_scope
    BEFORE INSERT OR UPDATE
    ON pack_users
    FOR EACH ROW
EXECUTE FUNCTION check_pack_user_role_scope();
