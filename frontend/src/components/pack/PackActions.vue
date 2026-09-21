<script setup lang="ts">
import {useSnackbarStore} from "@/stores/snackbar.ts";
import {type Pack, PackStatus} from "@/interfaces/pack.ts";
import {clientSetupCommandToClipboard, linkToClipboard, openPublicLink} from "@/services/packs.service.ts";

const {pack} = defineProps<{ pack: Pack }>()

const actionsDisabled = computed(() => {
  return pack.isArchived || pack.status === PackStatus.DRAFT
})

const snackbar = useSnackbarStore()

const copyToClipboard = async () => {
  await linkToClipboard(pack.id)
  snackbar.showSnackbar(
    'Link copied to clipboard',
    'default',
    2000
  )
}

const openLink = () => {
  openPublicLink(pack.id)
}

const copySetupCommand = async () => {
  await clientSetupCommandToClipboard(pack.id)
  snackbar.showSnackbar(
    'Client setup command copied to clipboard',
    'default',
    2000
  )
}

const actions = computed<{
  icon: string,
  action: () => void | Promise<void>,
  title: string,
}[]>(() => [
  {
    icon: 'mdi-clipboard-text-multiple-outline',
    action: copyToClipboard,
    title: pack.isPublic ? "Copy public link" : "Copy personalized link"
  },
  {
    icon: 'mdi-open-in-new',
    action: openLink,
    title: pack.isPublic ? "Open public link" : "Open personalized link",
  },
  {
    icon: 'mdi-console',
    action: copySetupCommand,
    title: "Copy client setup command (MultiMC/Prism)",
  }
])
</script>

<template>
  <v-tooltip
    v-if="actionsDisabled"
    text="Publish this pack to enable links"
    location="bottom"
  >
    <template #activator="{ props }">
      <div v-bind="props">
        <v-btn
          link
          density="comfortable"
          color="default"
          variant="plain"
          icon="mdi-dots-vertical"
          :disabled="true"
        />
      </div>
    </template>
  </v-tooltip>

  <v-menu v-else>
    <template #activator="{ props }">
      <v-btn
        link
        density="comfortable"
        color="default"
        variant="plain"
        v-bind="props"
        icon="mdi-dots-vertical"
        :disabled="actionsDisabled"
      />
    </template>

    <v-list>
      <v-list-item
        v-for="actionItem in actions"
        :key="actionItem.icon"
        :prepend-icon="actionItem.icon"
        :title="actionItem.title"
        @click="actionItem.action"
      />
    </v-list>
  </v-menu>
</template>
