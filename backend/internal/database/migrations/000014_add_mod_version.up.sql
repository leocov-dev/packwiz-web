-- Human-readable installed version of a mod (DB-only; never written to pack
-- files). NULL for rows created before this migration until their next update.
ALTER TABLE mods
    ADD COLUMN version TEXT;

-- Latest available version reported by the update check (may be NULL).
ALTER TABLE mod_update_checks
    ADD COLUMN latest_version TEXT;
