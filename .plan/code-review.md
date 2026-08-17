# Backend Code Review

Scope: `backend/` (Go). ~80 files, ~4.7k LOC. Read in full; findings below are
verified against actual library source (gorm v0.5+, gin v1.12.0) where behavior
was non-obvious, not assumed from general framework folklore.

Legend: 🔴 High impact · 🟠 Medium · 🟡 Low / cleanup

---

## 🔴 High-impact bugs

### 1. `AddMod` silently swallows the main mod's creation error and commits anyway
`internal/services/packwiz_svc/service.go:279-306`

```go
if err := ps.db.Transaction(func(tx *gorm.DB) error {
    ...
    for _, mod := range dependencies {
        dbMod, err := createMod(mod, dbPack, user, tx, true, nil)  // err scoped to loop body
        if err != nil { return err }
        ...
    }
    _, err = createMod(newMod, dbPack, user, tx, false, dependencyIds) // assigns outer `err`!
    return nil
}); err != nil {
    return response.Wrap(err)
}
```

`var err error` is declared at the top of `AddMod` (line 248). The `if err := ps.db.Transaction(...)`
on line 279 declares a *new*, if-scoped `err` — but that new variable isn't in scope yet while its
own initializer (the closure literal) is being evaluated, so the closure captures the **outer**
`AddMod`-level `err` by reference. Line 299's `_, err = createMod(newMod, ...)` writes into that
outer variable, not the transaction's return path, and the closure unconditionally `return nil`s
afterward.

Net effect: if creating the primary mod fails, the transaction still commits (dependency mods get
persisted, the primary mod does not), and `AddMod` returns `nil` — the API reports success on a
request that actually failed to add the requested mod. Confirmed via careful scope-tracing; the
loop-body `err` (line 292) is a distinct, correctly-checked variable — only the post-loop
assignment is broken.

**Fix**: use a fresh `:=` for the post-loop assignment and check it before falling through, e.g.
`if _, err := createMod(...); err != nil { return err }`.

### 2. `addModrinthMod` / `addCurseforgeMod` discard the dependency-lookup error
`internal/services/packwiz_svc/modrinth.go:60-78`, `internal/services/packwiz_svc/curseforge.go:55-74`

```go
missingDependencies, err := lookupModrinthDependencies(url, pack)
return mainMod, missingDependencies, nil   // <- `nil`, not `err`
```
Same shape in `addCurseforgeMod`. Both functions capture `err` from the dependency lookup and then
hardcode `nil` as the returned error, so a failed dependency lookup (network error, bad API
response) is invisible to the caller — it just looks like "no missing dependencies" instead of
"we don't know." Should be `return mainMod, missingDependencies, err`.

### 3. `ArchivePack` never un-publishes a public pack
`internal/services/packwiz_svc/service.go:346-368`

```go
tx.Model(&tables.Pack{ID: packId}).Updates(&tables.Pack{
    IsPublic: false,
    Status:   types.PackStatusDraft,
    DeletedAt: gorm.DeletedAt{Time: time.Now(), Valid: true},
})
```
GORM's struct-based `Updates()` only writes **non-zero** fields (documented behavior, not a
guess). `IsPublic: false` is the zero value for `bool`, so it is silently omitted from the
generated `UPDATE` — archiving a currently-public pack leaves `is_public = true` in the database.
`Status: types.PackStatusDraft` happens to work only because `"draft"` is a non-empty string.
Use `Select(...)` to force the field, or switch to a `map[string]interface{}` update.

### 4. `SetAcceptableVersions` targets a column name that doesn't exist
`internal/services/packwiz_svc/service.go:424-434`

```go
ps.db.Model(tables.Pack{}).Where(tables.Pack{ID: packId}).
    Update("acceptableGameVersions", request.Versions)
```
`Update(column, value)` uses the string literally as the SQL column name (no snake_case
translation the way struct-tag-driven updates get). The actual column, per migration
`000001_init_tables.up.sql:38`, is `acceptable_game_versions`. This call will fail with an
"undefined column" error the moment it's invoked. Currently **not called from anywhere**
(`SetAcceptableVersions` has no controller/route wiring) — dead code today, but a landmine for
whoever wires it up next, and it means the whole "acceptable versions" edit path was apparently
never manually tested end-to-end (`EditPack` sets the field on the model correctly via `Save`, so
that path is fine — this dedicated setter is the broken one).

---

## 🟠 Medium-impact issues

### 5. `GET /api/v1/admin/users` has no admin check
`internal/routes/admin.go`, `internal/server/router.go:69`

The admin route group is mounted under `protectedGroup` (which only requires
`middleware.ApiAuthentication`, i.e. "logged in") — there is no `IsAdmin` guard anywhere in the
chain. `tables.User.IsAdmin` is checked in exactly one place in the whole codebase
(`controllers/packwiz.go:63`, for pack creation). Any authenticated non-admin user can currently
page through the full user list (`GET /api/v1/admin/users`), including emails and full names.
Needs an `AdminGuard` middleware analogous to `PackPermissionGuard`, applied in
`RegisterAdminRoutes` or at the mount point in `router.go`.

### 6. Audit "action" tagging is wired to a key nothing reads
`internal/middleware/meta/audit_tags.go`, `internal/middleware/audit.go:61-65`

`meta.Tag(meta.CategoryLogin)` (used on the login route) sets context key `"meta.category"`.
`ApiAudit` only ever reads `"auditAction"`:
```go
if action := c.GetString("auditAction"); action != "" {
    auditRecord.Action = action
} else {
    auditRecord.Action = c.Request.Method + " " + c.FullPath()
}
```
The two never intersect, so every audit row's `Action` falls back to `"POST /api/v1/auth/login"`
etc. — the tagging feature (`meta.CategoryLogin` / `CategoryStatic`) is entirely inert. Either
have `ApiAudit` read `"meta.category"`, or remove the dead `meta` package.

### 7. Audit password redaction is dead code — and JSON bodies aren't redacted at all
`internal/middleware/audit.go:44-51`

```go
if err := c.Request.ParseForm(); err == nil {
    if len(c.Request.Form) != 0 {
        formCopy := utils.DeepCopyMapStringSlice(c.Request.Form)
        if _, ok := formCopy["password"]; ok {
            formCopy["password"] = []string{"********"}
        }
    }
}
```
`formCopy` is built, redacted, and then never assigned to `actionParams` or the audit record — the
whole block is a no-op (form data isn't captured in the audit log at all today, redacted or not).
Separately, JSON bodies *are* captured verbatim (`actionParams["body"] = bodyJson`, line 40) with
no redaction — currently harmless only because every endpoint carrying a password
(`LoginForm`, `ChangePasswordForm`) uses `form:` binding, not JSON. If a JSON-bound
password-bearing endpoint is ever added, it will be logged in plaintext into the `audit` table
with no code path currently guarding against it.

### 8. `EditPack` has inconsistent partial-update semantics
`internal/services/packwiz_svc/service.go:589-608`

```go
if request.Name != "" {
    pack.Name = request.Name
}
pack.Description = request.Description
pack.AcceptableGameVersions = request.AcceptableVersions
```
`Name` is only overwritten if non-empty (partial-update style), but `Description` and
`AcceptableGameVersions` are unconditionally overwritten — an edit request that omits them (or a
client that sends zero-values) silently wipes existing data. Worth deciding on one semantic
(full-replace vs. partial-patch) and applying it consistently; as written it looks like an
oversight rather than an intentional design choice.

### 9. `dto.AddModRequest.GitHub` is typed as `*AddCurseforge`
`internal/types/dto/add_mod.go:45-64`

```go
type AddGitHub struct {
    Url string `json:"url" validate:"required,url"`
}
...
type AddModRequest struct {
    Modrinth   *AddModrinth   `json:"modrinth"`
    Curseforge *AddCurseforge `json:"curseforge"`
    GitHub     *AddCurseforge `json:"github"`   // should be *AddGitHub
}
```
`AddGitHub` is defined and has its own `Validate()`/`IsSet()` but is dead — never referenced. It
works today only because `AddGitHub` and `AddCurseforge` happen to have identical shape; the two
types will silently drift apart the moment either one gains a field, at which point `AddModRequest`
would keep compiling against the wrong struct.

### 10. No mutual-exclusivity validation between mod-source fields
`internal/types/dto/add_mod.go:69-87`

`AddModRequest.Validate()` used to check "exactly one of modrinth/curseforge/github is set"
(the logic is present, commented out). As written, a request with **both** `modrinth` and
`curseforge` populated is accepted, and `AddMod`'s `if/else if` chain silently prefers Modrinth
with no signal to the caller that the other field was ignored. Worth restoring the check (updated
for the third `github` field) rather than leaving it commented out.

---

## 🟡 Low-impact / cleanup

### 11. `bootstrap debug` CLI command can never seed users
`commands/bootstrap.go`

```go
modeOptions = []string{"debug"}   // only "debug" passes arg validation
...
Run: func(cmd *cobra.Command, args []string) {
    choice := args[0]
    switch choice {
    case "users":                 // never matches "debug"
        seed.CreateRandomUsers(db, 50)
    }
},
```
The `Args` validator only accepts `"debug"` as the positional arg, but the `switch` only handles
`"users"` — `seed.CreateRandomUsers` is unreachable through the CLI as currently wired. Either the
`modeOptions` list or the `switch` case is stale from a refactor.

### 12. `ApiAuthentication` missing `return` after session-key-mismatch abort
`internal/middleware/authentication.go:35-40`

```go
if sessionKey == nil || sessionKey != user.SessionKey {
    ClearSession(c)
    log.Warn("session key mismatch")
    c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"msg": "session is invalid, log in again"})
}
c.Set("user", user)
c.Next()
```
No `return` after the abort branch. **Verified not exploitable**: `Abort()` sets
`c.index = abortIndex` (63), and Gin's `Next()` loop condition (`c.index < len(handlers)`) is
false for any realistic handler chain length, so the extra `c.Set`/`c.Next()` calls are no-ops and
no downstream handler actually runs — the 401 response already written is what the client sees.
Still worth adding the `return`: it's fragile (relies on an internal Gin implementation detail
staying true), wastes a `c.Set` call, and is a footgun if this pattern gets copy-pasted somewhere
with different control flow.

### 13. `PackwizModController`/`PackwizController` user-management endpoints are stubs
`internal/controllers/packwiz.go:320-334`

`GetPackUsers`, `AddPackUser`, `RemovePackUser`, `EditUserAccess` are all routed
(`routes/pack.go:41-44`, behind the edit-permission guard) but every handler just returns
`500 "not implemented"`. Not a bug, but these are live, permission-gated endpoints doing nothing —
worth a TODO/tracking issue or removing from the router until implemented, so client code doesn't
build against them.

### 14. `ListUsersQuery.NameSearch` / `EmailSearch` are bound but never used
`internal/types/dto/list_users_query.go`, `internal/services/user_svc/service.go:161-196`

Both fields are declared with `form:` tags and accepted by the query DTO, but
`UserService.ListUsers` only ever filters on `UserType` — search-by-name/email silently does
nothing. Likely an incomplete feature rather than a bug, but worth flagging since the API contract
implies filtering that isn't happening.

### 15. Leftover debug statement in query validation
`internal/types/dto/list_users_query.go:16-20`

```go
func (f *ListUsersQuery) Validate() error {
    log.Info(f)
    return validator.New(...).Struct(f)
}
```
Logs the struct pointer on every admin user-list request at `Info` level — looks like a debugging
leftover, not intentional operational logging.

### 16. `internal/services/importer/reconcile_dir.go` is 100% commented-out and stale
The entire file body is a comment block. It also references APIs that no longer exist in this
form (`packwiz_cli.GetPackFile`, `dr.packwizSvc.ArchivePack(slug string)` vs. the current
`ArchivePack(packId uint)`, `tables.PackUsers.PackSlug` vs. the current `PackID uint`) — this code
would not compile if uncommented. Either finish the importer or delete the dead file; as-is it's
misleading about the state of the import feature.

### 17. `SESSION_SECRET` has no minimum-strength enforcement, unlike `ADMIN_PASSWORD`
`internal/config/config.go:89-121`

`ADMIN_PASSWORD` panics at startup if unset or under 16 chars (`config.go:116-121`). `SESSION_SECRET`
defaults to the literal string `"insecure-session-secret"` with no equivalent check — a production
deployment that forgets to set `PWW_SESSION_SECRET` will silently sign session cookies with a
well-known, public value. Worth applying the same startup-panic treatment as the admin password.

### 18. `UpdateAll` / `UpdateMod` are unimplemented, matching TODO comments
`internal/services/packwiz_svc/service.go:436-513` — both correctly documented with `TODO: expose
in lib` and consistently return `501/500 "not implemented"`. Not a defect, listed for completeness
since they're reachable, permission-gated routes with no effect.

---

## Notes / non-findings (checked, ruled out)

- **Frontend static-file path traversal** (`controllers/frontend.go:42`): `filepath.Join("frontend",
  requestedPath)` looked suspicious, but `embed.FS.Open` validates against `fs.ValidPath` and
  rejects any path containing `..` segments — traversal attempts return an error rather than
  escaping the embedded root. No action needed.
- **SQL injection via `Search` in `GetPacksWithPerms`**: `"packs.slug LIKE ?", "%"+request.Search+"%"`
  concatenates into the bind *value*, not the query text — parameterized correctly, not
  injectable. (Unescaped `%`/`_` wildcards in user input can affect match breadth, but that's a
  UX nit, not a security issue.)
- **`ListUsers`'s discarded `query.Where(...)` return value** (`user_svc/service.go:171-176`):
  initially looked like the classic "forgot to reassign the GORM chain" bug, but traced through
  `gorm.DB.getInstance()`: `tx.Model(&tables.User{})` returns a `*DB` with `clone == 0` (the
  clone-tracking field isn't propagated into cloned instances), so the *subsequent* `.Where()`
  call mutates that same `*DB`'s `Statement` in place rather than cloning again — the filter is
  applied correctly despite the unused return value. Confirmed by reading `gorm.go`/
  `chainable_api.go` directly rather than assuming.

---

## Suggested priority order

1. Admin-route authorization gap (#5) — access-control issue, fix first.
2. `AddMod` swallowed-error/partial-commit bug (#1) — silent data corruption.
3. `ArchivePack` zero-value `Updates()` bug (#3) — visible, reproducible functional bug.
4. Dependency-lookup error discarding (#2) and audit-tag dead wiring (#6) — smaller blast radius,
   easy fixes.
5. Everything else in Medium/Low as time allows; several (11, 13, 16, 18) are really "finish or
   delete this stub" housekeeping rather than bugs.
