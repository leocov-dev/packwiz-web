# Project Plan

Tracking of outstanding features, improvements, and fixes across `frontend/` (Vue 3
+ Vuetify) and `backend/` (Go). Sourced from a full-repo review on 2026-08-16, plus a
survey of the `packwiz-nxt` library (`../packwiz-nxt`, `github.com/leocov-dev/packwiz-nxt`)
that the backend depends on for pack/mod operations. Completed items have been
removed as of 2026-08-31 — see git history of this file for the full record of what
shipped and when.

Legend: 🔴 High impact · 🟠 Medium · 🟡 Low / cleanup

**Maintenance note:** after completing any item from this doc, update it in the same
change — remove the item (or fold into a brief note if useful context), and adjust
the Suggested priority order below. Don't let this doc drift out of sync with the
actual code state.

## Features (missing or unimplemented)

### Modpack updating
- 🟡 **Mod re-resolution on MC/loader change is opt-in, not automatic.** The
  `Migrate` endpoint can cascade `core.UpdateAllMods` against the new target
  versions, but only when the caller opts in via the migrate dialog's
  checkbox — no automatic re-check, no background job, and no distinct
  "incompatible mod" flag/dry-run separate from actually updating. A mod with
  no compatible version for the new target just silently keeps its old
  `Download`/`Update` info rather than being surfaced as broken.
- No "update available" badge exists (`Mod.update` is a loosely-typed map,
  not a clean boolean flag) — would need a backend dry-run/check endpoint.
  Intentionally deferred, not a bug.

**Takeaway**: mod removal and Modrinth search have shipped; see git history of
this file. The remaining packwiz-nxt-derived polish items (bulk rehash,
optional mods, Quilt substitution, side-filtered listing) shipped 2026-08-31
— see below.

---

## Suggested priority order

Nothing outstanding. All previously-tracked packwiz-nxt-derived polish items
shipped 2026-08-31:
- Bulk rehash — `PATCH /pack/:packId/rehash`, backend's first use of
  `packwiz-nxt/fileio`, plus a "Rehash" dialog/button in `PackDetails.vue`.
- Optional mods — new `mods.option` JSONB column
  (migration `000004_add_mod_option`), `PATCH /pack/:packId/mod/:modId/option`,
  and Optional/Description/Default fields in `EditModForm.vue`.
- Quilt/Fabric dependency auto-substitution — verified already fully wired
  end-to-end (loader versions endpoint, DTO validation, `resolveLoaderVersion`,
  and `sources.MapDepOverride`'s `slices.Contains(..., "quilt")` gate all
  already supported it); no code change needed.
- Side-filtered mod listing — frontend-only filter (`lib/mod-filters.ts` +
  a side `v-select` in `ModsList.vue`), since mods are already fully loaded
  client-side with the pack; no new backend endpoint.
