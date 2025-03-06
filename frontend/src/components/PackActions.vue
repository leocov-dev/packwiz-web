<script setup lang="ts">
import {useSnackbarStore} from "@/stores/snackbar.ts";
import {type Pack, PackStatus} from "@/interfaces/pack.ts";
import {linkToClipboard, openPublicLink} from "@/services/packs.service.ts";

const {pack} = defineProps<{ pack: Pack }>()

const snackbar = useSnackbarStore()


const copyToClipboard = async () => {
  await linkToClipboard(pack.slug)
  snackbar.showSnackbar(
    'Link copied to clipboard',
    'default',
    2000
  )
}

const openLink = () => {
  openPublicLink(pack.slug)
}

const actions: {
  icon: string,
  action: () => void | Promise<void>,
  tooltip: string,
}[] = [
  {
    icon: 'mdi-clipboard-text-multiple-outline',
    action: copyToClipboard,
    tooltip: pack.isPublic ? "Copy public link" : "Copy personalized link"
  },
  {
    icon: 'mdi-open-in-new',
    action: openLink,
    tooltip: pack.isPublic ? "Open public link" : "Open personalized link",
  }
]
</script>

<template>
  <div>
    <template
      v-for="actionItem in actions"
      :key="actionItem.icon"
    >
      <v-tooltip :text="actionItem.tooltip">
        <template #activator="{ props }">
          <v-btn
            link
            density="comfortable"
            color="default"
            variant="plain"
            v-bind="props"
            :icon="actionItem.icon"
            :disabled="pack.isArchived || pack.status === PackStatus.DRAFT"
            @click="actionItem.action"
          />
        </template>
      </v-tooltip>
    </template>
  </div>
</template>

