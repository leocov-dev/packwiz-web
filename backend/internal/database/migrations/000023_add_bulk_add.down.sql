UPDATE pack_snapshots SET reason = 'mod_add' WHERE reason = 'mod_bulk_add';
ALTER TABLE pack_snapshots DROP CONSTRAINT pack_snapshots_reason_check;
ALTER TABLE pack_snapshots ADD CONSTRAINT pack_snapshots_reason_check CHECK (reason IN (
    'baseline', 'publish', 'pack_edit',
    'mod_add', 'mod_remove', 'mod_update',
    'mod_side', 'mod_option', 'mod_pin', 'mod_uplift',
    'rehash', 'update_all', 'migrate',
    'migrate_mods'));

DROP TABLE pack_bulk_add_items;
DROP TABLE pack_bulk_add_runs;
