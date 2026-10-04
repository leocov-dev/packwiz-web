<script setup lang="ts">

import {useAuthStore} from "@/stores/auth";
import {useAppStore} from "@/stores/app";
import {Perm, type PermissionName} from "@/lib/permissions.ts";
import {usePermissions} from "@/composables/usePermissions.ts";
import {SOURCE_REPO_URL} from "@/lib/links.ts";

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
  {
    text: 'Pack Access',
    icon: 'mdi-chart-bar',
    route: '/admin/pack-access',
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

    <div class="d-flex align-center justify-center ga-2 text-caption text-medium-emphasis pb-2 px-1">
      <a
        :href="SOURCE_REPO_URL"
        target="_blank"
        rel="noopener noreferrer"
        class="d-inline-flex text-medium-emphasis text-decoration-none"
        aria-label="Source code on GitHub"
        title="Source code on GitHub"
      >
        <v-icon
          icon="mdi-github"
          size="20"
        />
      </a>
      <span
        v-if="appStore.version"
        class="text-truncate"
      >v{{ appStore.version }}</span>
    </div>
  </v-list>
</template>

