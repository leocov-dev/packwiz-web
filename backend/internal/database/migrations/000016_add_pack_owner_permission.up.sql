-- Owner (30) is reserved for the pack creator. Promote the creator's existing
-- access row, or create it if the creator was somehow removed.
INSERT INTO pack_users (pack_id, user_id, permission)
SELECT id, created_by, 30
FROM packs
ON CONFLICT (pack_id, user_id) DO UPDATE SET permission = 30;
