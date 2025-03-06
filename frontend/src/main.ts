/**
 * main.ts
 *
 * Bootstraps Vuetify and other plugins then mounts the App`
 */
import "reflect-metadata";

// Plugins
import { registerPlugins } from '@/plugins'

// Components
import App from './App.vue'

// Composables
import { createApp } from 'vue'
import router from '@/router'
import {initializeCacheStore} from "@/stores/cache.ts";

const app = createApp(App)

registerPlugins(app)

Promise.all([
  router.isReady(),
  initializeCacheStore(),
]).then(() => {
  app.mount('#app')
})
