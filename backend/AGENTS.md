# AGENTS.md — packwiz-web/backend

Guidance for anyone (human or agent) writing Go code in this module. This is
the API/backend for the packwiz-web SPA: a Gin HTTP server backed by GORM,
using `packwiz-nxt` for packwiz domain logic. Follow the existing layering
strictly — `controllers → services → database/tables` — and keep each layer's
responsibilities separate.

## 1. Layering & responsibilities

- **`internal/controllers`**: HTTP-only concerns. Bind/validate the request
  (`mustBindJson`, `mustBindQuery`, `mustBindIdParam`, `mustBindCurrentUser`,
  etc. from `internal/controllers/utils.go`), call exactly one service method,
  translate the result/error into a response (`isOK`, `dataOK`,
  `abortWithError`). **No business logic, no direct DB/GORM calls, no direct
  filesystem or `packwiz-nxt` calls in a controller.** If a handler is doing
  more than bind → call service → respond, that logic belongs in a service.
- **`internal/services/*_svc`**: all business/domain logic. Each service is a
  small struct constructed with `NewXxxService(db *gorm.DB, ...) *XxxService`
  (see `packwiz_svc`, `user_svc`, `auth_svc`) and takes its dependencies
  (`*gorm.DB`, other services) as constructor parameters — never as package
  globals. New services must follow this same constructor-injection shape so
  they stay testable and swappable.
- **`internal/database`/`internal/tables`**: schema, migrations, and GORM
  model definitions. Query logic that's reused across services belongs here
  or in the owning service, not copy-pasted between controllers.
- **`internal/config`**: the only place allowed to read environment variables
  or `viper` config. Every other package must receive configuration through
  `config.C` or, better, through explicit parameters/struct fields passed in
  at construction time — don't reach for `os.Getenv`/`viper` from a
  controller or service.

## 2. Error handling convention

- Controllers use `response.ServerError` (see
  `internal/types/response/http_error.go`) as the error type returned up
  through bind helpers and service calls: `response.New(code, message)` for a
  specific HTTP status/message, `response.Wrap(err)` for an unexpected
  internal error (defaults to `500`). Every controller method that can fail
  should return/short-circuit via `pc.abortWithError(c, err)` immediately
  after the call that can fail — don't collect errors and check them later.
- Services should return plain Go `error` (wrap with `fmt.Errorf("...: %w",
  err)` when adding context), and let the controller decide the HTTP status
  via `response.Wrap`/`response.New`. Don't import `gin` or construct HTTP
  responses from inside a service.
- Never swallow an error silently (`_ = someCall()`) unless it is genuinely
  safe to ignore (e.g. best-effort cleanup) — and if so, comment why.

## 3. SOLID in this codebase

- **SRP**: one controller struct per resource (`PackwizModController`, etc.),
  one service per bounded concern (`packwiz_svc`, `user_svc`, `auth_svc`).
  Don't let a controller grow request-shaping/business logic, and don't let a
  service reach into `gin.Context`.
- **OCP**: prefer adding a new controller method / service method over
  branching deeply inside an existing one when adding a new capability;
  extend via new small methods rather than widening existing ones with flags.
- **LSP**: `response.ServerError` implementations (`HttpError` and any future
  ones) must all honor the interface contract (`Error() string`,
  `JSON(c *gin.Context)`) consistently — a `JSON` implementation that behaves
  differently per type (e.g. skips setting the status code) breaks every
  caller that relies on `abortWithError`.
- **ISP**: keep interfaces like `dto.Request` (bind + `Validate()`) and
  `response.ServerError` small and focused; don't bolt unrelated methods onto
  them for one caller's convenience.
- **DIP**: services depend on `*gorm.DB` and other services passed in via
  constructor, not on globals; controllers depend on service *types*
  constructed once (typically in `internal/server`/route setup) and injected,
  not re-constructed per request unless required (e.g. per-request state).

## 4. Go idioms

- Run `make fmt` (gofmt, includes `go mod tidy`) before committing; check
  with `make fmtcheck` if unsure.
- Prefer explicit constructors (`NewXxx(...)`) over exported structs with
  public fields the caller must remember to initialize, following the
  existing `*_svc` pattern.
- Keep DTOs (`internal/types/dto`) as the request/response boundary; don't
  pass GORM table models (`internal/tables`) directly to/from the HTTP layer
  — map explicitly, as `ListMissingDependencies` does in
  `internal/controllers/packwiz-mod.go`.
- Validate all inbound request DTOs via their `Validate()` method as part of
  binding (see `mustBindJson`/`mustBindQuery`/`mustBindForm`) — don't add a
  second, ad-hoc validation path in the controller body.
- Use `context`-aware GORM calls where the underlying method supports it, and
  don't hold long-lived goroutines/background work without a way to cancel
  them.
- Guard secrets and required configuration the way `internal/config/config.go`
  does: fail fast (`panic` at startup, not silently default) when a required
  security-relevant setting (e.g. `ADMIN_PASSWORD`) is missing or invalid;
  this is the one place in the module where an early panic is intentional and
  acceptable — request-handling code should never panic.
- Keep doc comments on exported identifiers, starting with the identifier's
  name, per standard Go convention.

## 5. Testing

- Add/extend tests alongside new service logic; business rules belong in
  services, so that's where most test coverage should live (controllers are
  thin enough that they mainly need coverage for bind/validation edge cases).
- Run `make test` before submitting a change; it runs the full suite with
  coverage and must pass.

## 6. Before submitting a change

1. `make fmt` (or `make fmtcheck` to verify without modifying).
2. `make test` — all existing tests must still pass; add new tests for new
   behavior.
3. Re-read your diff against sections 1–3: did a controller gain business
   logic or a direct DB call, did a service reach into `gin.Context` or read
   env vars directly, or did an error get swallowed instead of
   wrapped/returned?

## 7. OIDC login

Multi-provider OpenID Connect lives in `internal/services/oidc_svc` (provider
CRUD and client cache, the flow, and the pure decision function), with thin
controllers `oidc_auth.go`, `oidc_identity.go` and `oidc_admin.go`, and routes in
`internal/routes/oidc.go`. Identities are stored in `user_identities`; user-side
helpers (`CreateExternalUser`, `LinkIdentity`, `UnlinkIdentity`, `SetPassword`)
are in `user_svc/identity.go`. Setup for humans is in `docs/oidc.md`. Rules to
keep when changing it:

- **Resolution order** is in `oidc_svc.Decide` (pure, table-tested): existing
  identity (`provider_id + sub`), then verified-email link if `link_by_email`,
  then auto-create if `auto_create_users`, else reject. Never match on email
  alone. Keep new rules inside `Decide` so they stay testable.
- **The superuser (`admin`) is untouchable** from OIDC paths: it cannot be created,
  linked, modified or deactivated, and auto-created users only get the `user` role.
- **State cookie**: the flow state (state, nonce, PKCE verifier, mode, link user
  id, redirect) is sealed with `utils.SecretboxEncrypt` in a `SameSite=Lax`
  cookie, because the `Strict` session cookie is not sent on the cross-site
  callback. The callback must never read the session; link mode takes the user
  id from the sealed state captured in `Begin`. Clear the cookie on every callback.
- **Redirects**: only same-origin relative paths are allowed (`redirect.go`).
  Failures redirect with a fixed error code; never forward IdP error text.
- **Secrets**: the client secret is encrypted at rest (key derived from
  `SESSION_SECRET`), write-only in the API (`hasSecret` only, blank on update
  keeps it), and redacted in audit params. Provider create/update is refused
  while `config.DefaultSessionSecret` is in use. A secret that no longer decrypts
  marks the provider broken; it must never crash the app or block local login.
- **Audit**: the public login/callback routes have no user, so audit would drop
  them (callback uses `SkipAudit`). Authenticated link/unlink routes are tagged
  `meta.CategoryOidcLink` / `meta.CategoryOidcUnlink`.
- **Rate limiting** for these routes is delegated to the edge proxy, not the app.
- Tests use `httptest` for a fake IdP (discovery, JWKS, token endpoint); there
  is no DB test harness, so keep logic in pure functions.

## 8. Authorization (RBAC)

Every access check flows through `permission <- role_permissions <- role`; plan
in `.plan/rbac.md`. Rules to keep:

- **Never check a role or the old admin flag in code.** Decisions live only in
  `internal/services/authz_svc` (`DecideGlobal`, `DecidePack`, `CanOnPack`). The
  bootstrapped superuser (`users.is_superuser`) short-circuits there and nowhere else.
- **Guards**: protect routes with `middleware.RequirePermission(name)` (global)
  or `middleware.RequirePackPermission(authz, name)` (pack). Permission names are
  constants in `authz_svc/permissions.go` and must match the seeding migrations
  (000017, 000018; `matrix_test.go` replays them in order).
- **Archived packs** are read-only for everyone, superuser included; only the
  permissions in `archivedAllowed` pass (409 otherwise).
- **Owner identity** is `packs.created_by`, not a role. The `owner` role is not assignable.
- **New routes** need a permission guard, or an entry in `authOnlyRoutes` in
  `internal/routes/guard_coverage_test.go`. That test fails otherwise.
- **New permissions or roles**: add the migration rows, the constant, and update
  `AllPermissions`, `globalPermissions` and `expectedMatrix` in
  `authz_svc/matrix_test.go`; the test checks them against the migration SQL.

## 9. Pack history (snapshots)

A published pack keeps a full snapshot of its content after every change (tables
`pack_snapshots`, head pointer `packs.head_snapshot_id`). Code in
`packwiz_svc/history*.go`. Rules to keep:

- **Every write to pack content or mods goes through `withPackHistory`**, which locks
  the pack row first (lock order is always pack, then mods), applies the change, then
  records the snapshot in the same transaction. If the snapshot fails the change rolls
  back; history is not optional. A new mutator that skips it silently breaks revert.
  Writes that are not content (users, roles, status, `is_public`, archive) do not snapshot.
- **Drafts and archived packs record nothing.** Publishing records a snapshot if the
  content differs from the head. A published pack with no head gets a `baseline`
  snapshot before its first change (`ensureHistoryBaseline`).
- **Snapshots are never deleted.** A revert marks the live snapshots after its target
  `abandoned` and moves the head; it writes no new snapshot. Abandoned snapshots can be
  viewed and cloned, not restored.
- **Mods are stored by slug**, dependencies as slugs, so a revert keeps row ids (and
  `mod_update_checks`) for mods that still exist. `mods.version` is display-only: it is
  stored but excluded from the content hash.
- **Never `Save()` a pack loaded with `Mods` preloaded** (it re-inserts mods and writes
  back stale status/public/deleted_at). Use `Select(cols).Updates(...)`.
  `packs.head_snapshot_id` is read-only to gorm; only `setPackHead` writes it.
- **Snapshot payload changes** need a new `historySchemaVersion` and a branch in
  `decodeHistoryPayload`. Keep payload build, hash, diff and restore planning in pure
  functions (`history_payload.go`); there is no DB test harness.
