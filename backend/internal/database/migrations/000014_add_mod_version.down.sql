ALTER TABLE mod_update_checks
    DROP COLUMN IF EXISTS latest_version;
ALTER TABLE mods
    DROP COLUMN IF EXISTS version;
