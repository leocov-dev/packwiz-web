# AGENTS.md — packwiz-web/frontend

Guidance for anyone (human or agent) writing Vue/TypeScript code in this
module. This is the SPA frontend for packwiz-web: Vue 3 (`<script setup>`
composition API) + Vuetify 3 + Pinia + Vue Router (file-based, via
`unplugin-vue-router`), talking to the `packwiz-web/backend` API through a
thin Axios service layer. Follow the existing layering — `pages/components →
stores/composables → services → interfaces` — and keep each layer's
responsibilities separate, mirroring the discipline already documented in
`packwiz-web/backend/AGENTS.md`.

## 1. Layering & responsibilities

- **`src/pages`**: route-level components only (file-based routing via
  `unplugin-vue-router` — the file path *is* the route). A page composes
  layout + components + a store/composable call; it should not itself
  contain API calls or business logic. Layouts live in `src/layouts` and are
  applied via `unplugin-vue-layouts`.
- **`src/components`**: presentational/interactive building blocks, grouped
  by feature (`components/mods`, `components/forms`, `components/pack`,
  `components/user`). A component receives data via `defineProps`, emits
  intent via `defineEmits`/`defineModel`, and calls a **service** function
  directly for the one action it owns (see `AddModForm.vue` calling
  `addMod`/`listMissingDependencies`) — it does not reach into `apiClient`
  directly, bypassing `src/services`.
- **`src/stores`** (Pinia): cross-cutting/shared application state
  (`auth`, `user`, `cache`, `snackbar`). Use the typed `defineStore<Id, State,
  Getters, Actions>(...)` form with explicit `StateInterface`/`ActionsInterface`
  /`GettersInterface` types, as every existing store does — don't fall back to
  untyped stores for new state. Actions are the only place a store talks to a
  service; components/pages call store actions, not services directly, for
  anything that must be shared/cached across the app (auth state, cached
  version/loader data). Purely local, single-component state does **not**
  need a store — use local `ref`/`reactive` instead.
- **`src/composables`**: reusable *stateless-per-call* reactive logic not
  tied to a single global store (see `buildDataLoader` in
  `composables/data-loader.ts` for the loading/data/error pattern). Prefer a
  composable over duplicating the same `isLoading`/`error` `ref` boilerplate
  across components.
- **`src/services`**: the only layer allowed to import `apiClient`
  (`services/api.service.ts`) and know API routes/HTTP verbs. Each function
  is a thin, named async wrapper (`fetchOneMod`, `addMod`,
  `listMissingDependencies`) that takes plain arguments, calls `apiClient`,
  and returns typed data (via `plainToInstance` for response DTOs). No Vue
  reactivity, no component/store logic here.
- **`src/interfaces`**: response/request DTOs. Use `class-transformer`
  classes (with `@Type(() => X)` for nested class hydration) for anything
  that flows through `plainToInstance`, as `interfaces/pack.ts` does; use
  plain `interface`/`type` for request-only shapes that never need
  hydration (`interfaces/requests.ts`). Don't reuse a Pinia store's internal
  `interface` as a wire DTO or vice versa — keep the API boundary and store
  shape independently defined even if they overlap in fields.

## 2. SOLID in this codebase

- **SRP**: one component = one visual/interactive concern. If a component
  is doing data-fetching *and* form validation *and* submission *and*
  navigation, extract pieces into a composable or push the fetch into a
  service call the component merely awaits (see how `AddModForm.vue` keeps
  URL-parsing (`parseUrl`), request-building (`buildRequest`), and
  submission (`submitForm`) as separate, individually named functions rather
  than one large handler).
- **OCP**: add new mod-source support, new pack fields, or new pages by
  adding a new component/route/service function, not by widening an
  existing function with another `if` branch on a type discriminator unless
  that discriminator is genuinely small and closed (e.g. `parseUrl`'s
  source detection is acceptable because the set of sources is fixed and
  small; don't extend that pattern to open-ended cases).
- **LSP**: components that accept the same prop "shape" for interchangeable
  use (e.g. any component rendering a `Mod`) must handle every valid value
  of that type consistently — don't special-case one variant's fields in a
  way that breaks for another variant of the same declared type.
- **ISP**: keep `defineProps<{...}>()`/`defineEmits<{...}>()` contracts
  minimal — pass only what a component needs (`ModCard` takes `packId` and
  `mod`, not the whole `Pack`). Don't grow a shared interface (e.g. `Mod`,
  `Pack`) with fields only one consumer needs; add an optional field only
  when it's genuinely part of that domain object.
- **DIP**: components/pages depend on **services** and **store actions**
  (abstractions over "how data is fetched/cached"), never on `axios`/
  `apiClient` directly (the only exception is `api.service.ts` itself and
  code explicitly checking `axios.isAxiosError` for error narrowing).
  Stores depend on services, not the reverse; services depend on
  `interfaces`, not on Vue/Pinia.

## 3. Go idioms → Vue/TypeScript idioms

- **Composition API only.** All new components use `<script setup lang="ts">`;
  do not introduce Options API components.
- **Typed props/models.** Use `defineProps<{...}>()` / `defineModel<T>(...)`
  with explicit TypeScript types (see `SlugAndName.vue`, `ModCard.vue`).
  Avoid the runtime `defineProps({...})` object form and avoid `any` — if a
  prop's type isn't naturally expressible, define an `interface`/`type` in
  `src/interfaces` and import it.
- **Reactive props destructuring** (Vue 3.5+) is fine and used throughout
  (`const {packId, mod} = defineProps<{...}>()`) — prefer it over
  `props.foo` access for readability, but remember destructured props stay
  reactive only through the compiler macro; don't manually cache a
  destructured prop into a plain local `let`/`const` if it needs to update.
- **Rely on auto-import, but only for what it covers.** `ref`, `computed`,
  `watch`, `onMounted`, `useRouter`, etc. are auto-imported (see
  `unplugin-auto-import`/`unplugin-vue-components` config) and used without
  explicit `import` statements inside `.vue` files — don't add redundant
  `import {ref} from 'vue'` there. Plain `.ts` files (composables, stores,
  services) still need explicit imports, as `data-loader.ts` does.
- **Path alias**: always import project modules via the `@/` alias
  (`@/services/...`, `@/interfaces/...`, `@/stores/...`), never relative
  `../../` chains.
- **Error handling**: wrap awaited service/API calls in `try/catch/finally`,
  toggle a local `loading`/`error` ref, and narrow Axios errors explicitly
  with `axios.isAxiosError(e)` before reading `e.response?.data` (see
  `AddModForm.vue`, `AuthStore.changePassword`). Don't let a rejected
  promise propagate unhandled out of a component's event handler.
- **No `any`.** Prefer precise types; where a Pinia getters interface
  requires an index signature (framework limitation), keep the
  `eslint-disable-next-line @typescript-eslint/no-explicit-any` comment
  scoped to that single line, as existing stores do — don't spread `any`
  further than that.
- **Naming conventions**: components `PascalCase.vue`; composables
  `camelCase.ts` exporting a `useXxx`/`buildXxx` function; stores
  `camelCase.ts` exporting `useXxxStore`; services `kebab-case.service.ts`
  (or `camelCase.service.ts` matching existing files) exporting verb-named
  functions (`fetchOneMod`, `addMod`); interfaces/DTOs `camelCase.ts`
  exporting `PascalCase` classes/types.
- Run `npm run lint` (ESLint, `--fix`) and `npm run type-check` (`vue-tsc`)
  before committing — both must be clean.

## 4. Vuetify/UI conventions

- Build UI from Vuetify components (`v-card`, `v-text-field`, `v-form`,
  `v-btn`, etc.) rather than hand-rolled markup/CSS where an equivalent
  Vuetify component exists; use Vuetify's spacing/utility classes (`ma-`,
  `pa-`, `d-flex`, etc.) instead of ad-hoc `<style>` blocks.
- Form validation: define a local `rules` object of small named validator
  functions (see `SlugAndName.vue`, `AddModForm.vue`) and pass them via
  `:rules="[...]"`; bind form validity to a `v-model` `isValid` ref on
  `v-form` and gate submission on it.
- Keep template logic simple (conditionals on already-computed refs/getters);
  push any non-trivial branching/computation into the `<script setup>` block
  as a named function or `computed`, not inline in the template.

## 5. Before submitting a change

1. `npm run lint` — ESLint clean (auto-fixed where possible).
2. `npm run type-check` — `vue-tsc` must report no errors.
3. `npm run build` if the change touches build-affecting config (env,
   plugins, routing) — must succeed.
4. Re-read your diff against sections 1–2 above: did a component reach past
   `services`/store actions into `apiClient` directly, did business logic
   leak into a page/component that belongs in a service or store action, or
   did a shared interface/DTO get widened for one caller's convenience?
