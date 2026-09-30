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
      <div
        v-bind="props"
        class="d-flex align-center"
        tabindex="0"
        role="group"
        aria-label="Links unavailable: publish this pack to enable links"
      >
        <v-btn
          v-for="actionItem in actions"
          :key="actionItem.icon"
          density="comfortable"
          variant="text"
          :icon="actionItem.icon"
          :aria-label="actionItem.title"
          disabled
        />
      </div>
    </template>
  </v-tooltip>

  <div
    v-else
    class="d-flex align-center"
  >
    <v-tooltip
      v-for="actionItem in actions"
      :key="actionItem.icon"
      :text="actionItem.title"
      location="bottom"
    >
      <template #activator="{ props }">
        <v-btn
          v-bind="props"
          density="comfortable"
          variant="text"
          :icon="actionItem.icon"
          :aria-label="actionItem.title"
          @click="actionItem.action"
        />
      </template>
    </v-tooltip>
  </div>
</template>
