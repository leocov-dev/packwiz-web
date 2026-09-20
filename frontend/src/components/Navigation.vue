<script setup lang="ts">

import {useAuthStore} from "@/stores/auth";
import {useAppStore} from "@/stores/app";

const route = useRoute()
const authStore = useAuthStore()
const appStore = useAppStore()
const selected = ref<string[]>([])

const items = [
  {
    text: 'Mod Packs',
    icon: 'mdi-package-variant',
    route: '/packs',
  },
  {
    text: 'Users',
    icon: 'mdi-account-group',
    route: '/admin/users',
  },
  {
    text: 'Audit Log',
    icon: 'mdi-format-list-text',
    route: '/admin/audit',
  },
]

const userItems = computed(() => {
  return items.filter(item => {
    return !(item.route.startsWith('/admin') && (!authStore.user || !authStore.user.isAdmin))
  })
})

watch(
  () => route.path,
  (newPath) => {
    selected.value = [newPath]
  },
  { immediate: true }
)

</script>

<template>
  <v-list
    v-model:selected="selected"
    select-strategy="single-leaf"
    nav
    class="d-flex flex-column fill-height"
  >
    <v-list-item
      v-for="(item, i) in userItems"
      :key="i"
      :value="item.route"
      :to="item.route"
      :prepend-icon="item.icon"
      :title="item.text"
      color="primary"
    />

    <v-spacer />

    <div
      v-if="appStore.version"
      class="text-caption text-medium-emphasis text-center text-truncate pb-2 px-1"
    >
      v{{ appStore.version }}
    </div>
  </v-list>
</template>

