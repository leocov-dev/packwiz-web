<script setup lang="ts">
import {useSnackbarStore} from "@/stores/snackbar.ts";
import {type Pack, PackStatus} from "@/interfaces/pack.ts";
import {linkToClipboard, openPublicLink} from "@/services/packs.service.ts";

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
  <v-tooltip
    v-if="actionsDisabled"
    text="Publish this pack to enable links"
    location="bottom"
  >
    <template #activator="{ props }">
      <div v-bind="props">
        <template
          v-for="actionItem in actions"
          :key="actionItem.icon"
        >
          <v-btn
            link
            density="comfortable"
            color="default"
            variant="plain"
            :icon="actionItem.icon"
            :disabled="true"
          />
        </template>
      </div>
    </template>
  </v-tooltip>

  <div v-else>
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
            :disabled="actionsDisabled"
            @click="actionItem.action"
          />
        </template>
      </v-tooltip>
    </template>
  </div>
</template>

