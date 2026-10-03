<script setup lang="ts">

import {useAuthStore} from "@/stores/auth";
import {useAppStore} from "@/stores/app";
import {Perm, type PermissionName} from "@/lib/permissions.ts";
import {usePermissions} from "@/composables/usePermissions.ts";

const route = useRoute()
const authStore = useAuthStore()
const {can} = usePermissions()
const appStore = useAppStore()
const selected = ref<string[]>([])

const items: { text: string, icon: string, route: string, permission?: PermissionName }[] = [
  {
    text: 'Mod Packs',
    icon: 'mdi-package-variant',
    route: '/packs',
  },
  {
    text: 'Users',
    icon: 'mdi-account-group',
    route: '/admin/users',
    permission: Perm.UserView,
  },
  {
    text: 'OIDC Providers',
    icon: 'mdi-key-chain',
    route: '/admin/oidc',
    permission: Perm.OidcManage,
  },
  {
    text: 'Audit Log',
    icon: 'mdi-format-list-text',
    route: '/admin/audit',
    permission: Perm.AuditView,
  },
]

const userItems = computed(() => {
  return items.filter(item => {
    return !item.permission || (!!authStore.user && can(item.permission))
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

