-- Prune stray (abandoned) snapshots and rebase a pack's history onto a chosen snapshot.
-- Both permanently delete snapshot rows.

INSERT INTO permissions (name, resource, description)
VALUES ('pack.snapshot.manage', 'pack', 'Prune abandoned snapshots and rebase the pack history');

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
         JOIN permissions p ON (
    (r.name = 'system_admin' AND p.name IN ('pack.snapshot.manage'))
        OR (r.name = 'owner' AND p.name IN ('pack.snapshot.manage'))
    );
