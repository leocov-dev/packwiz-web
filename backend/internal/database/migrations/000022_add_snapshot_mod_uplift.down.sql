UPDATE pack_snapshots SET reason = 'mod_pin' WHERE reason = 'mod_uplift';
ALTER TABLE pack_snapshots DROP CONSTRAINT pack_snapshots_reason_check;
ALTER TABLE pack_snapshots ADD CONSTRAINT pack_snapshots_reason_check CHECK (reason IN (
    'baseline', 'publish', 'pack_edit',
    'mod_add', 'mod_remove', 'mod_update',
    'mod_side', 'mod_option', 'mod_pin',
    'rehash', 'update_all', 'migrate',
    'migrate_mods'));
