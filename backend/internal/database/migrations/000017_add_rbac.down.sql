-- Static (permission = 1) grants deleted by the up migration cannot be restored.

DROP TRIGGER trg_pack_users_scope ON pack_users;
DROP FUNCTION check_pack_user_role_scope();

ALTER TABLE pack_users
    ADD COLUMN permission SMALLINT NOT NULL DEFAULT 1 CHECK ( permission BETWEEN 0 AND 999 );

UPDATE pack_users pu
SET permission = CASE r.name WHEN 'viewer' THEN 10 WHEN 'editor' THEN 20 WHEN 'owner' THEN 30 ELSE 10 END
FROM roles r
WHERE r.id = pu.role_id;

DROP INDEX idx_pack_users_role_id;
ALTER TABLE pack_users
    DROP COLUMN role_id;

ALTER TABLE users
    ADD COLUMN is_admin BOOLEAN NOT NULL DEFAULT FALSE;

UPDATE users u
SET is_admin = TRUE
WHERE u.is_superuser
   OR EXISTS (SELECT 1
              FROM user_roles ur
                       JOIN roles r ON r.id = ur.role_id
              WHERE ur.user_id = u.id
                AND r.name = 'system_admin');

DROP INDEX idx_users_single_superuser;
ALTER TABLE users
    DROP COLUMN is_superuser;

DROP TABLE user_roles;
DROP TABLE role_permissions;
DROP TABLE permissions;
DROP TABLE roles;

DROP FUNCTION check_user_role_scope();
DROP FUNCTION check_role_permission_scope();
