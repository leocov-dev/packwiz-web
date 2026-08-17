# Project Plan

Consolidated tracking of features, improvements, and fixes across `frontend/` (Vue 3
+ Vuetify) and `backend/` (Go). Sourced from a full-repo review on 2026-08-16, plus a
survey of the `packwiz-nxt` library (`../packwiz-nxt`, `github.com/leocov-dev/packwiz-nxt`)
that the backend depends on for pack/mod operations. Re-verified against backend code
on 2026-08-17 after commit `47c34ff` ("fix: address backend code review findings") and
further uncommitted backend work landed — see **Update log** below.

`.plan/code-review.md` (the original line-level bug writeup this doc's Fixes section
was based on) has since been deleted from the working tree; its findings are folded
into this doc and the git history of `47c34ff` still has the full text if needed.

Legend: 🔴 High impact · 🟠 Medium · 🟡 Low / cleanup

## Update log

**2026-08-17 — backend re-scan.** Two things changed the picture since the
2026-08-16 review:

1. Commit `47c34ff` fixed essentially the entire "Fixes" section below (backend
   bugs #1–#12, #14–#17; #8's `EditPack` partial-update semantics too). All are now
   verified fixed by reading current code — details moved to "Fixes ✅ resolved"
   below. Only the *feature* gaps they were adjacent to (mod update, MC/loader
   migration, pack-collaborator UI) remain open.
2. **Uncommitted, currently in the working tree** (`git status`): the backend halves
   of two previously-"stub" features have been implemented:
   - `UpdateAll`/`UpdateMod` (`packwiz_svc/service.go`) now actually call
     `core.UpdateAllMods`/`core.UpdateSingleMod` and persist results — no longer
     `501 not implemented`.
   - `GetPackUsers`/`AddPackUser`/`RemovePackUser`/`EditUserAccess`
     (`packwiz.go:325-405`, backed by new `PackwizService` methods and a new
     `dto/pack_user.go`) are now fully implemented — no longer stubs.
   - Audit middleware also picked up JSON-body redaction (`utils.RedactSensitiveValues`)
     to go with the pre-existing form redaction.

   **Both of these are backend-only** — the frontend still has zero UI or service
   calls for either. See updated sections below.

---

## Features (missing or unimplemented)

### Modpack updating
- 🔴 **Pack Minecraft/loader version edit doesn't persist — still true.**
  `PackEditForm.vue` submits a `minecraft`/`loader` block; backend `EditPack`
  (`backend/internal/services/packwiz_svc/service.go:751-775`, current as of
  2026-08-17) still only reads `Name`/`Description`/`AcceptableVersions` — no
  `MinecraftDef`/`LoaderDef` handling exists. Not touched by the recent fix commit.
  The form appears to succeed but silently no-ops.
- ✅ **Update mod(s) to latest version — backend now implemented (uncommitted).**
  `UpdateAll`/`UpdateMod` (`backend/internal/services/packwiz_svc/service.go:442-479,
  538-573`) now call `core.UpdateAllMods`/`core.UpdateSingleMod` from packwiz-nxt and
  persist the result transactionally via a new `applyModUpdate` helper (writes
  `FileName`/`Download`/`Source`/`Update`/`UpdatedBy` back onto `tables.Mod`).
  `UpdateMod` still correctly rejects pinned mods
  (`400 "cannot update pinned mod"`); `UpdateAll` relies on `UpdateAllMods` to skip
  pinned mods internally. Both controller handlers now thread the current user
  through for the new `UpdatedBy` field. **This flips the remaining work to
  frontend-only**: no service call, button, "update available" badge, or bulk
  "update all" action exists anywhere in the UI — routes `PATCH pack/:id/update-all`
  and `PATCH mod/:id/update` are fully live and unused.
- 🔴 **No re-resolution of mod compatibility when MC/loader version changes.**
  Version-aware mod resolution (`modrinth.go`/`curseforge.go` `ModrinthGetLatestVersion`
  etc.) only fires once, at initial add-mod time. There's no bulk re-check, no
  incompatibility flag, no background job. Moot in practice today since (1) above
  means the pack's stored MC version can't even be changed.
- Suggested build order (updated): backend mod-update work is done, so this is now
  → fix `EditPack` MC/loader persistence → build frontend UI (per-mod update
  button, bulk "update all", "update available" badges using the existing
  `Mod.update` field, `frontend/src/interfaces/pack.ts:62`, wired to the two
  already-working endpoints).

### MC version / loader migration (library already supports it — backend never called it)
- `EditPack`'s no-op for `MinecraftDef`/`LoaderDef` (see above) isn't a missing
  library feature — `packwiz-nxt` has a full `migrate minecraft`/`migrate loader`
  implementation (`internal/commands/cmdmigrate/{minecraft,loader}.go`) that the
  backend simply never calls:
  - `core.GetMinecraftVersions()` + `McVersionInfo.CheckValid(version)`
    (`core/mcversion.go:35, 31`) — full valid-MC-version list and validation.
  - `core.ModLoaders` map (`core/versionutil.go:21`) covering fabric, forge,
    liteloader, quilt, neoforge, each with a version-list getter per MC version —
    this is what a "pick a loader + loader version" dropdown would be built on.
  - `core.GetForgeRecommended(mcVersion)` — Forge's separate
    "recommended" vs. "latest" version concept.
  - The CLI's migrate flow also supports an optional cascading "update mods to
    match the new MC version" step, which would naturally chain into the mod-update
    feature above.
  - `core.PackToml.AddAcceptableVersion`/`RemoveAcceptableVersion`/
    `SetAcceptableGameVersions` — incremental add/remove of acceptable MC versions,
    vs. backend's current all-or-nothing overwrite via the (currently broken,
    see Fixes #4) `SetAcceptableVersions`.

### Mod management
- 🔴 **Edit Mod is a non-functional stub.** `frontend/src/components/mods/EditModForm.vue:20`
  — submit handler has the real API call commented out
  (`// await addMod(pack.id, request)`); no fields are bound; no `editMod`/`updateMod`
  function exists in `frontend/src/services/mods.service.ts`. Button is even
  mislabeled "Add Mod".
- 🟠 **No mod removal/delete UI or backend call wired from the frontend.**
  `ModCard.vue` only has an Edit button.
- 🟡 Commented-out "open source link" button in
  `frontend/src/components/mods/ModCard.vue:16-18, 34-43`.

### Admin section (frontend entirely placeholder; backend is a mixed bag per sub-feature)
- 🔴 All four admin routes render only `<UnderConstruction />`, yet are linked
  directly from primary nav for admin users (dead-end from main UI):
  - `frontend/src/pages/admin/audit/index.vue:11`
  - `frontend/src/pages/admin/users/index.vue:11`
  - `frontend/src/pages/admin/users/[id].vue:15`
  - `frontend/src/pages/admin/users/[id].edit.vue:15`
  - `frontend/src/pages/admin/users/new.vue:11`
  - Nav links: `frontend/src/components/Navigation.vue:12-21`
- 🟠 `frontend/src/components/user/UserList.vue` is an orphaned, unfinished
  copy-paste of `PackList.vue` — not referenced from any page, `reload` is a no-op,
  links to a nonexistent `/users/new` route, renders `PackCard` instead of a user
  card. Its filter config offers "Show Active"/"Show Deactivated" checkboxes, but
  **`tables.User` has no active/deactivated concept at all** (only GORM soft-delete
  via `DeletedAt`, which no admin action currently sets) — the filter UI doesn't
  correspond to any real backend state, so this component can't just be "reconnected",
  it needs the underlying user-lifecycle model decided first.
- 🟠 No corresponding `user.service.ts` functions for admin user CRUD or audit log
  fetching.
- **Backend readiness per admin sub-feature (checked 2026-08-17):**
  - ✅ **List users** — `GET /api/v1/admin/users` (`AdminController.GetUsersPaginated`
    → `UserService.ListUsers`) is fully implemented: paginated, `nameSearch`/
    `emailSearch`/`userType` filters all work (the `#14` bug where search filters
    were bound-but-unused is fixed), and the route is now correctly gated by
    `middleware.AdminGuard` (`#5` fixed). Zero frontend usage today — pure wiring
    task.
  - 🟠 **Get single user by ID** — missing. No `GET /admin/users/:id` route or
    controller method exists; `/admin/users/[id].vue` has nothing to call.
  - 🟠 **Edit an arbitrary user (admin-initiated)** — partially there.
    `UserService.UpdateUser(userId, dto.EditUserRequest)` already takes a target
    `userId` and would work unmodified, but it's currently only reachable through
    `UserController.UpdateUser` (`POST /user/update`), which is hardcoded to
    `mustBindCurrentUser(c)` — i.e. self-service only. Needs a new admin-scoped
    controller method + route (`PATCH /admin/users/:id` or similar) that accepts an
    admin-supplied ID; the service layer needs no changes.
  - 🔴 **Create user** — missing entirely. No `POST /admin/users` (or any user
    creation) endpoint anywhere in the backend. The only way a `tables.User` row
    gets created today is the `bootstrap` CLI command seeding an initial admin.
    `/admin/users/new.vue` has nothing to build against yet.
  - 🔴 **Deactivate/reactivate user** — missing, and blocked on the same
    active/deactivated-model gap noted above for `UserList.vue`. Needs a schema/
    design decision (soft-delete vs. a dedicated `IsActive` flag) before either end
    can be built.
  - 🟠 **Audit log — write path is solid, read path doesn't exist.**
    `middleware.ApiAudit` (`internal/middleware/audit.go`) already persists a
    `tables.Audit` row per authenticated request, and just improved further
    (uncommitted): it now reads the `meta.category` tag set via
    `meta.Tag(meta.CategoryLogin)` (`#6` fixed) and redacts sensitive values in both
    form data (`password`) and arbitrary JSON bodies via the new
    `utils.RedactSensitiveValues` (`#7` fixed, and hardened further than the
    original fix — now covers any key containing "password"/"secret"/"token", not
    just literal `password`). **But there is no read side at all**: no
    `AuditService`, no controller method, no route — `internal/routes/admin.go`
    only registers `GET admin/users`. Building the audit page needs, from scratch:
    a query/list method (with pagination — `tables.Audit` rows accumulate on every
    request), a controller, an `AdminGuard`-protected route, and then the frontend
    service + page.

### Pack-level collaborator/permission management
- ✅ **Backend now fully implemented (uncommitted).**
  `GetPackUsers`/`AddPackUser`/`RemovePackUser`/`EditUserAccess`
  (`backend/internal/controllers/packwiz.go:325-405`, routed in
  `routes/pack.go:41-44`) are no longer stubs. Backed by new `PackwizService`
  methods (`ListPackUsers`, `GrantPackUser`, `RevokePackUser`,
  `ChangePackUserPermission` — `packwiz_svc/service.go:644-748`) with sensible
  validation: grant checks the pack and user both exist and rejects a duplicate
  grant with `409`; revoke/edit both `404` if the user has no existing access. New
  `dto/pack_user.go` validates `Permission` against the real
  `PackPermissionStatic/View/Edit` enum. `ListPackUsers` joins to `users` for
  username/full name/email so a "manage collaborators" table can be built directly
  off the response shape (`PackUserInfo`).
- 🟠 **Frontend has nothing** — no `packs.service.ts` functions, no component, no
  entry point in `PackDetails.vue`/`PackEditForm.vue` — despite `PackPermission`
  enum (`frontend/src/interfaces/pack.ts`) already being checked and displayed
  (`PackCard.vue`, `PackDetails.vue`). This is now a pure frontend build: add
  service calls for the four endpoints, and a collaborator-management UI (likely a
  dialog or tab off pack edit) — no backend work needed.

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
- 🟠 **Search Modrinth by name/keyword.** `sources.ModrinthSearchForProjects(query, versions)`
  (`sources/mr-api.go:50`) — full-text project search filtered by MC version.
  Today's add-mod flow only accepts a direct URL paste
  (`modrinthProjectAndVersion`/`curseforgeModInfoFromUrl`/`addGithubMod` all take a
  `url` string) — there is no "search for a mod by name" anywhere in the product,
  even though Modrinth search is already available in the library. (No equivalent
  free-text CurseForge search function was found — only slug/URL/ID lookups via
  `sources.CurseforgeModInfoFromSlug`.)
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
  target loader anywhere in the (currently broken) MC-version/loader edit flow, so
  it's dead in practice until loader migration (above) is built.
- 🟡 **Side-filtered listing semantics.** The CLI's `list` command has defined
  fallthrough rules for `ServerSide`/`ClientSide`/`UniversalSide`/`EmptySide`
  filtering (`cmd/list.go`) that aren't replicated in any packwiz-web listing/filter
  endpoint — the frontend's mod list has no side-based filter at all currently.

**Takeaway**: the two highest-value items — mod updates and MC/loader migration —
are folded into the "Modpack updating" feature section above since the library
already does the hard part; this section covers the rest (export, import, search,
external-URL mods, rehash, optional mods) which are complete, ready-to-call library
features with zero footprint anywhere in packwiz-web today.

### Minor / scale
- 🟡 No pagination in `PackList.vue`/`ModsList.vue` (`items-per-page="0"`, load-all)
  — may be fine at current scale, revisit if pack/mod counts grow.

---

## Fixes (bugs in existing functionality)

### Backend — ✅ resolved (verified against current code, 2026-08-17)
All of the following were fixed by commit `47c34ff` ("fix: address backend code
review findings"). Re-read the current source for each to confirm — all confirmed
fixed, none regressed:

- ✅ **#5 Admin routes had no admin-role check.** `routes.RegisterAdminRoutes` is now
  called with `middleware.AdminGuard(db)` (`internal/server/router.go:69`).
- ✅ **#1 `AddMod` swallowed the primary mod's creation error and committed anyway.**
  Now `if _, err := createMod(newMod, ...); err != nil { return err }`
  (`packwiz_svc/service.go:298-300`).
- ✅ **#3 `ArchivePack` never un-published a public pack.** Now uses
  `.Select("IsPublic", "Status", "DeletedAt")` before `.Updates(...)`
  (`packwiz_svc/service.go:350-365`).
- ✅ **#4 `SetAcceptableVersions` targeted a nonexistent column name.** Now
  `Update("acceptable_game_versions", ...)` (`packwiz_svc/service.go:428-434`).
- ✅ **#2 `addModrinthMod`/`addCurseforgeMod` discarded the dependency-lookup error**
  — propagated now.
- ✅ **#6 Audit "action" tagging was wired to a key nothing read.** `ApiAudit` now
  checks `c.Get("meta.category")` first, falling back to `auditAction`
  (`internal/middleware/audit.go:63-66`).
- ✅ **#7 Audit password redaction was dead code; JSON bodies weren't redacted.**
  Fixed, then hardened further (uncommitted, 2026-08-17): JSON bodies now go
  through `utils.RedactSensitiveValues`, which redacts any key containing
  "password"/"secret"/"token" (`internal/utils/structures.go`), not just a literal
  `password` key.
- ✅ **#8 `EditPack` had inconsistent partial-update semantics.**
  `Description`/`AcceptableVersions` now only overwrite when non-empty, matching
  `Name`'s existing behavior (`packwiz_svc/service.go:758-768`). Note: this is a
  separate issue from the still-open MC/loader-version persistence gap — see
  Features → Modpack updating.
- ✅ **#9 `dto.AddModRequest.GitHub` was typed as `*AddCurseforge` instead of
  `*AddGitHub`** — corrected, with mutual-exclusivity validation restored (#10).
- ✅ **#10 No mutual-exclusivity validation between mod-source fields** — restored.
- ✅ **#11 `bootstrap debug` CLI command could never seed users** — arg/switch
  mismatch fixed.
- ✅ **#12 `ApiAuthentication` missing `return` after session-key-mismatch abort** —
  defensive `return` added.
- ✅ **#13 Pack-user-management endpoints were stubs** — now fully implemented, see
  Features → Pack-level collaborator/permission management above (this is the
  "backend done, frontend still missing" item).
- ✅ **#14 `ListUsersQuery.NameSearch`/`EmailSearch` bound but never used** — now
  applied in `UserService.ListUsers`.
- ✅ **#15 Leftover debug `log.Info(f)` in `ListUsersQuery.Validate()`** — removed.
- ✅ **#16 `internal/services/importer/reconcile_dir.go` fully commented-out/stale**
  — deleted.
- ✅ **#17 `SESSION_SECRET` had no minimum-strength enforcement** — same startup
  panic guard as `ADMIN_PASSWORD` added.
- ✅ **#18 `UpdateAll`/`UpdateMod` unimplemented** — now implemented (uncommitted),
  see Features → Modpack updating above.

**Nothing outstanding in this section.** The only backend items still open are
feature-shaped, not bug-shaped, and are tracked in Features above: `EditPack`
MC/loader persistence, admin single-user-get/create/deactivate endpoints, and the
audit-log read path.

### Frontend
- 🟡 `frontend/src/components/forms/MinecraftVersion.vue:11` — TODO: "Latest"/"Latest
  Snapshot" formatted version strings aren't valid to submit to the backend as-is;
  `PackEditForm.vue` special-cases stripping them back out, moot until the `EditPack`
  persistence fix above lands.
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

*(Superseded 2026-08-17 — all backend bug fixes from the original #1-#5 landed;
re-ranked around what's actually left.)*

1. **Commit the in-flight backend work.** `UpdateAll`/`UpdateMod` and the four
   pack-user endpoints are fully implemented but sitting uncommitted in the working
   tree — get these landed before building frontend against them.
2. **Frontend: wire mod updates** — `PATCH pack/:id/update-all` and
   `PATCH mod/:id/update` are live and unused. Per-mod update button, bulk "update
   all", "update available" badge off `Mod.update`. No backend work needed.
3. **Frontend: pack-collaborator management UI** — `GetPackUsers`/`AddPackUser`/
   `RemovePackUser`/`EditUserAccess` are live and unused. Add `packs.service.ts`
   calls + a manage-collaborators screen/dialog. No backend work needed.
4. **Frontend: Edit Mod stub** — smallest, highest-value UI fix; currently misleads
   users into thinking an edit succeeded.
5. **Backend: `EditPack` MC/loader-version persistence + MC/loader migration** —
   `core.GetMinecraftVersions`/`core.ModLoaders`/`GetForgeRecommended` already exist
   in packwiz-nxt, the backend just never calls them. Unblocks a UI flow that's
   currently silently broken.
6. **Admin user management — pick a scope and build it end-to-end.** Backend is
   uneven per sub-feature (list: done; get-by-id/create/deactivate: missing;
   edit-arbitrary-user: service method exists, needs a controller+route). Decide
   which of create/edit/deactivate are actually needed before building, since
   deactivate needs a schema decision first (soft-delete vs. `IsActive` flag).
7. **Audit log read path** — write path is solid and just got better (JSON
   redaction); needs an `AuditService` query method, controller, `AdminGuard`
   route, and frontend from scratch. No existing scaffolding to build on, unlike
   the other admin sub-features.
8. Decide scope on: packwiz export/import (ready-to-call packwiz-nxt library
   functions, see the packwiz-nxt section above) — larger feature builds, not
   quick fixes.
9. Everything else in Medium/Low as time allows.
