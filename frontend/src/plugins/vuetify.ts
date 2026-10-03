/**
 * plugins/vuetify.ts
 *
 * Framework documentation: https://vuetifyjs.com`
 */

// Styles
import '@mdi/font/css/materialdesignicons.css'
import 'vuetify/styles'

// Composables
import {createVuetify} from 'vuetify'
import {md3} from 'vuetify/blueprints'
import {ThemePackwiz} from "@/themes/theme-packwiz.ts";

// https://vuetifyjs.com/en/introduction/why-vuetify/#feature-guides
export default createVuetify({
  blueprint: md3,
  // One surface radius app-wide: matches VCard from the md3 blueprint ('lg').
  defaults: {
    VSheet: {rounded: 'lg'},
    VToolbar: {rounded: 'lg'},
    // App bar is full-width chrome, not a surface.
    VAppBar: {rounded: 0, VToolbar: {rounded: 0}},
    VAlert: {rounded: 'lg'},
  },
  theme: {
    defaultTheme: 'dark',
    themes: {
      light: ThemePackwiz.light,
      dark: ThemePackwiz.dark,
    }
  },
})
