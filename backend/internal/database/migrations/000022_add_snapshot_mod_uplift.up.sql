-- Allow the 'mod_uplift' snapshot reason (dependency promoted to a regular mod).
ALTER TABLE pack_snapshots DROP CONSTRAINT pack_snapshots_reason_check;
ALTER TABLE pack_snapshots ADD CONSTRAINT pack_snapshots_reason_check CHECK (reason IN (
    'baseline', 'publish', 'pack_edit',
    'mod_add', 'mod_remove', 'mod_update',
    'mod_side', 'mod_option', 'mod_pin', 'mod_uplift',
    'rehash', 'update_all', 'migrate',
    'migrate_mods'));
