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

## Outstanding verification gap

A large batch of backend + frontend work (MC/loader migration, admin user
lifecycle including the `IsActive` migration, audit log, pack-collaborator
management, Update All) has **not been live-verified in a browser or against a
real Postgres** in this environment — Docker Desktop here couldn't bind-mount
`/Users/leo/src` or forward the local Postgres port
(`localdev/docker-compose.yml`'s `55432:5432` unreachable from host, though
`docker exec pg_isready` succeeds inside the container). Worth a real
end-to-end check next time the local stack is reachable, especially the
`000003_add_user_is_active` migration.

---

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

### Mod management
- ✅ **Mod removal/delete UI shipped 2026-08-31.** `ModCard.vue` gained a
  delete button (confirm dialog, error banner) calling the already-implemented
  `DELETE v1/packwiz/pack/:packId/mod/:modId` endpoint; both Edit and Delete
  are now gated on the `canEdit` prop, which is newly threaded through from
  `ModsList.vue` (previously ungated).
- 🟡 **`RemoveModById` doesn't clean up stale `DependencyIds` references.**
  Deleting a mod that other mods depend on leaves their `DependencyIds`
  arrays pointing at a nonexistent mod row
  (`backend/internal/services/packwiz_svc/service.go:525`,
  `backend/internal/tables/mod.go`'s `DependencyIds` field). Pre-existing gap,
  not introduced by the removal UI. Needs a product decision: cascade-delete
  dependents, block deletion of a depended-upon mod, or filter stale IDs at
  read time.
- 🟡 Commented-out "open source link" button in
  `frontend/src/components/mods/ModCard.vue:16-18, 34-43`.

### Packwiz file import
- 🟡 Backend `internal/services/importer/reconcile_dir.go` is 100% commented-out,
  stale, and references APIs that no longer exist in this form (would not compile if
  uncommented). No corresponding frontend UI at all. Decide: finish or delete.
- See also the packwiz-nxt import capabilities below — the stale reconciler predates
  what the library can now do directly.

### Available in packwiz-nxt but not exposed anywhere in packwiz-web
The backend only touches a thin slice of `packwiz-nxt`: add-mod resolution
(modrinth/curseforge/github by URL) and missing-dependency lookup. The library
itself is a full library-ified fork of the original `packwiz` CLI (`init`, `list`,
`pin`/`unpin`, `refresh`, `rehash`, `remove`, `update`, `migrate minecraft`/
`migrate loader`, `settings acceptable-versions`, `modrinth {install,export}`,
`curseforge {install,export,import,detect}`, `github install`, `url add`) and
supports considerably more than what's wired up. None of the below has any backend
route or frontend UI today:

- 🟠 **Export to `.mrpack` (Modrinth pack format).**
  `sources.BuildModrinthManifest` (`sources/mr-export.go:56`) +
  `sources.CanBeIncludedDirectly` (`sources/mr-export.go:30`). No export capability
  of any kind exists in packwiz-web currently.
- 🟠 **Export to CurseForge modpack zip.** `sources.ParseExportData`
  (`sources/cf-updater.go:423`) + `internal/commands/cmdcurseforge/packinterop.WriteManifestFromPack`
  — builds `manifest.json` + `modlist.html` + overrides.
- 🟠 **Import a whole CurseForge modpack.** `sources.CurseforgeImportPack`
  (`sources/cf-import.go:19`) — resolves every CF mod reference from a
  `manifest.json`/`minecraftinstance.json`, a zip, or a URL to one, into a
  `core.Pack`. Handles zip/dir/URL/Windows-Curse-instance source detection. This is
  a complete, ready-to-call replacement for the stale `reconcile_dir.go` importer
  above.
- 🟡 **Bulk-detect mods from a folder of jars.** `sources.CurseforgeDetectMods(dir)`
  (`sources/cf-detect.go:47`) — hashes `.jar`/`.litemod` files with CurseForge's
  murmur2 fingerprint and matches them against the CF API. Could back a "drop in an
  existing mods folder, tell me what's in it" import flow.
  - Note: no `.mrpack` *import* exists in the library (export only) — would need
    building even at the library layer if wanted.
- ✅ **Search Modrinth by name/keyword — shipped 2026-08-31.** New
  `PackwizService.SearchModrinthProjects` (`packwiz_svc/modrinth.go`) wraps
  `sources.ModrinthSearchForProjects`, mapped to a `dto.ModSearchResult` DTO
  rather than leaking the third-party Modrinth type. New
  `GET pack/:packId/mod/search` route/controller
  (`SearchModrinthMods` in `packwiz-mod.go`), inheriting the edit-permission
  guard. Frontend: `AddModForm.vue` gained a "Paste URL"/"Search Modrinth"
  toggle — selecting a search result populates the same `modUrl` field the
  existing URL-paste flow already consumes, so add/dependency-check/submit
  logic is unchanged. No CurseForge equivalent exists in the library (only
  slug/URL/ID lookups via `sources.CurseforgeModInfoFromSlug`), so CF search
  remains out of scope.
- 🟡 **Add a mod from a direct download URL (non-provider "external" mod).** The
  CLI's `url add` command builds a `core.ModToml` with `Download.Mode = core.ModeURL`
  by hashing the file via `fileio.CreateDownloadSession`
  (`internal/commands/cmdurl/install.go`) — the underlying data model
  (`core.ModDownload{URL, HashFormat, Hash}`) already supports this, but there's no
  exported `sources.*` helper for it (logic is inlined in the CLI command) and
  packwiz-web only supports modrinth/curseforge/github as mod sources today.
- 🟡 **Bulk rehash.** The `rehash` command (`cmd/rehash.go`) re-computes and rewrites
  every mod's hash to a chosen format (sha1/sha256/sha512) via
  `fileio.CreateDownloadSession`. No equivalent in the backend.
- 🟡 **Optional mods (per-mod default on/off toggle).** `core.ModOption{Optional bool,
  Description string, Default bool}` (`core/modtoml.go:48`), exposed on
  `core.Mod.Option`. The backend already wires up `ChangeModSide` (client/server/
  universal, see `ModsList.vue` grouping) but nothing sets or reads `Option` — the
  "mark a mod optional with a default state" concept is entirely unused.
- 🟡 **Quilt/Fabric dependency auto-substitution** — `sources.MapDepOverride`
  (`sources/depoverride.go`, used from `sources/cf-updater.go:540`) automatically
  swaps Fabric-only deps (Fabric API, Fabric Language Kotlin) for Quilt-native
  equivalents when the pack targets Quilt. This fires implicitly today inside the
  already-used `*FindMissingDependencies` calls, but Quilt isn't selectable as a
  target loader anywhere in the MC-version/loader edit flow, so it's dead in
  practice until Quilt is added as a selectable loader.
- 🟡 **Side-filtered listing semantics.** The CLI's `list` command has defined
  fallthrough rules for `ServerSide`/`ClientSide`/`UniversalSide`/`EmptySide`
  filtering (`cmd/list.go`) that aren't replicated in any packwiz-web listing/filter
  endpoint — the frontend's mod list has no side-based filter at all currently.

**Takeaway**: mod removal and the packwiz-nxt-backed features (export, import,
search, external-URL mods, rehash, optional mods) are the largest remaining gaps —
complete, ready-to-call library features with zero footprint anywhere in
packwiz-web today.

### Minor / scale
- 🟡 No pagination in `PackList.vue`/`ModsList.vue` (`items-per-page="0"`, load-all)
  — may be fine at current scale, revisit if pack/mod counts grow.

---

## Fixes (bugs in existing functionality)

### Frontend
- 🟡 `frontend/src/components/forms/MinecraftVersion.vue:11` — TODO: "Latest"/"Latest
  Snapshot" formatted version strings aren't valid to submit to the backend as-is;
  `PackEditForm.vue` special-cases stripping them back out.
- 🟡 Commented-out manual refresh button in
  `frontend/src/components/pack/PackDetails.vue:171-176`.
- 🟡 Debug-level `console.debug` calls left in `frontend/src/router/index.ts:19,22`
  and `frontend/src/stores/auth.ts:54,59`.
- 🟡 `frontend/README.md` is empty — no frontend-specific setup/dev instructions.

---

## Test coverage (gap, not a bug)
- No test runner configured anywhere in `frontend/` (`package.json` has no
  vitest/jest/cypress/playwright) and zero `*.spec.*`/`*.test.*` files in the repo.
  Notable given non-trivial client logic (URL-source parsing in `AddModForm.vue`,
  filter/query-string sync, auth guard logic).

---

## Suggested priority order

1. Verify the batch of recently-shipped work end-to-end against a real Postgres
   (see Outstanding verification gap above) — especially the `IsActive` migration.
2. Mod removal/delete UI.
3. Decide scope on: packwiz export/import (ready-to-call packwiz-nxt library
   functions, see above) — larger feature builds, not quick fixes.
4. Everything else in Medium/Low as time allows.
