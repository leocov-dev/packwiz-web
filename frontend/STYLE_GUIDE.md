# Frontend style guide

Rules for UI in `packwiz-web/frontend`. Goal: look like stock Vuetify 3 (MD3
blueprint) everywhere. Reuse Vuetify components, props, theme colors and
utility classes. Do not invent styles.

## Order of preference

1. Vuetify component + props (`variant`, `color`, `density`, `size`, `rounded`).
2. Vuetify utility classes (`ma-*`, `pa-*`, `ga-*`, `d-flex`, `text-*`, `cursor-pointer`, `text-break`, ...).
3. Theme colors and tokens (see Theme). No custom colors unless needed or requested, and recorded.
4. Custom CSS. Last resort, see below.

## Theme

Defined in `src/themes/theme-packwiz.ts`, registered in `src/plugins/vuetify.ts`
on top of the Vuetify `md3` blueprint. Two themes, `light` and `dark`. The
default is `dark`. Users pick `system`, `dark` or `light` (`ThemeSwitch.vue`,
`composables/userTheme.ts`). Never switch themes any other way.

### Brand colors

Only these are overridden. Same value in both themes.

| Color | Value | Use |
| --- | --- | --- |
| `primary` | `#8b3eaa` | Brand purple. Primary action, selected state, links, dependency relationship. |
| `secondary` | `#665593` | Muted purple. Rare: decorative or supporting surfaces. Not for buttons or chips. |

### Semantic colors

Not overridden, so Vuetify defaults apply and adapt per theme.

| Color | Meaning |
| --- | --- |
| `success` | Completed, healthy, update available |
| `warning` | Needs attention, admin, pinned notices |
| `error` | Failed, deactivated, destructive action |
| `info` | Neutral notice, optional |

### Surfaces and text

Use theme tokens through Vuetify props and classes, never raw colors:

- Backgrounds: `background` (app), `surface` (cards, dialogs, list items).
  Classes: `bg-surface`, `bg-transparent`.
- Text: default `on-surface`; secondary text `text-medium-emphasis`;
  inactive text `text-disabled`; brand or semantic text `text-primary`,
  `text-error`, etc.
- Shades: `primary-darken-1..4`, `primary-lighten-1..4` exist. Use sparingly,
  only when a token can't do it (e.g. `UserMenu.vue` list background).

### Custom colors

Do not use custom colors (hex, rgb, hsl, new theme colors, or shades outside
the tokens above) unless one is specifically needed or the user asks for it.
When one is used:

- Mark it in the code with a comment starting `custom-color:` that says why
  and who asked, e.g. `<!-- custom-color: brand badge, requested by user -->`
  or `/* custom-color: ... */`.
- Add it to the list below so it can be tracked and reviewed.

Current custom colors: none.

### Changing the theme

- Change or add a brand color only in `theme-packwiz.ts`, in both `light` and
  `dark`, then check both.
- To add a color, add it to the theme, then use it by name (`color="x"`).
  Do not put a hex/rgb value in a component or stylesheet (see Custom colors).
- Check contrast in both themes: text on `primary` must stay readable.

## Buttons

| Role | Props | Notes |
| --- | --- | --- |
| Primary | `color="primary"` | Flat is the default. One per view or dialog. |
| Secondary | `variant="tonal"` | Optional `color` only when semantic. |
| Tertiary / Cancel | `variant="text"` | No `color`. |
| Icon only | `icon`, `variant="text"`, `density="comfortable"` | Always set `aria-label`. |
| Destructive | `color="error"` | Flat for the final confirm, `text`/`tonal` elsewhere. |

- No `variant="outlined"` or `variant="plain"` on buttons.
- No `color="warning"`, `secondary`, `surface-variant` or `success` for ordinary actions.
- Use `density="comfortable"` for buttons inside rows and cards, default elsewhere.
- Use the `text` prop for labels, `prepend-icon` for a leading icon.

## Chips and badges

Base: `size="small" label variant="tonal"`.

- No color for neutral metadata (source, side, tags).
- Color only carries meaning: `success` (ok, update available), `warning`
  (attention, admin), `error` (failed, deactivated), `info` (optional),
  `primary` (relationship or role, e.g. dependency, admin).
- Use `prepend-icon` for an icon. Do not use `x-small`, `outlined` or `flat`.
- Filter chips: `<v-chip-group>` with `filter label variant="tonal"`.

## Form fields

Use the blueprint defaults. Do not set `variant` on fields, except `solo` for
fields inside a `v-toolbar`. Use `density="compact"` only in toolbars and rows.

## Layout and surfaces

- Pages are `v-card` on the app background. List items are `v-card`
  (`elevation-4`, like `ModCard`) or `v-list-item`.
- Border radius: `lg` on every surface (`v-card`, `v-sheet`, `v-toolbar`,
  `v-alert`), set once in `src/plugins/vuetify.ts`. Don't set `rounded` on them.
  The app bar (`v-app-bar`) is square.
- Page content sits `ma-6` from the main view edge. Put it on the page's root
  element (or the list component in the page), never flush to the edge.
- Alerts: `v-alert` with `type`; use `variant="tonal"` for inline hints.
- Spacing comes from utility classes only. No pixel values.

## Custom CSS

Avoid. Before adding any `<style>` block or `style=""` attribute, confirm no
Vuetify prop or utility class does the job. If you add one, put a comment on
it saying why none applies.

Sanctioned exceptions (each has a comment in the code):

- `PackCard.vue` `.multiline-truncate`: multi-line clamp, no utility.
- `ModCard.vue` `.mod-version`, `.mod-version-prefix`: width cap and
  visually-hidden text, no utility.
- `EditModForm.vue` visually-hidden status text: no utility.
- `AuditList.vue` `max-width` on truncated params: no utility.

Never use `!important`, unrecorded custom colors or `:deep()` restyling of Vuetify
internals.
