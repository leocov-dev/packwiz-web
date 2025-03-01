<script setup lang="ts">

import {useAuthStore} from "@/stores/auth";

const authStore = useAuthStore()

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
</script>

<template>
  <v-list
    nav
    class="d-flex flex-column fill-height"
  >
    <v-list-item
      v-for="(item, i) in userItems"
      :key="i"
      :value="item"
      :to="item.route"
      :prepend-icon="item.icon"
      :title="item.text"
      color="primary"
    />
  </v-list>
</template>

