<script setup lang="ts">
import type {Pack} from "@/interfaces/pack.ts";
import {
  clientSetupCommandToClipboard,
  downloadMultiMCInstance,
  instanceUrlToClipboard,
  linkToClipboard,
  openPublicLink,
} from "@/services/packs.service.ts";
import {useSnackbarStore} from "@/stores/snackbar.ts";

const {pack} = defineProps<{ pack: Pack }>()
const model = defineModel<boolean>({required: true})

const snackbar = useSnackbarStore()

const linkKind = computed(() => pack.isPublic ? "public" : "personalized")

const copyLink = async () => {
  await linkToClipboard(pack.id)
  snackbar.showSnackbar('Link copied to clipboard', 'default', 2000)
}

const openLink = () => {
  openPublicLink(pack.id)
}

const downloading = ref(false)
const downloadInstance = async () => {
  downloading.value = true
  try {
    await downloadMultiMCInstance(pack.id)
  } catch (e) {
    snackbar.showSnackbar((e as Error).message, 'error', 3000)
  } finally {
    downloading.value = false
  }
}

const copyInstanceUrl = async () => {
  await instanceUrlToClipboard(pack.id)
  snackbar.showSnackbar('Import URL copied to clipboard', 'default', 2000)
}

const copySetupCommand = async () => {
  await clientSetupCommandToClipboard(pack.id)
  snackbar.showSnackbar('Client setup command copied to clipboard', 'default', 2000)
}
</script>

<template>
  <v-dialog
    v-model="model"
    max-width="600"
  >
    <v-card title="Pack links and setup">
      <v-card-text>
        <h4 class="mb-1">
          Pack link
        </h4>
        <p class="text-medium-emphasis mb-3">
          The {{ linkKind }} link to this pack's metadata. Use it with any packwiz installer.
        </p>
        <div class="d-flex ga-2 mb-6">
          <v-btn
            variant="tonal"
            density="comfortable"
            prepend-icon="mdi-clipboard-text-multiple-outline"
            text="Copy link"
            @click="copyLink"
          />
          <v-btn
            variant="tonal"
            density="comfortable"
            prepend-icon="mdi-open-in-new"
            text="Open link"
            @click="openLink"
          />
        </div>

        <h4 class="mb-1">
          MultiMC or Prism Launcher instance
        </h4>
        <p class="text-medium-emphasis mb-3">
          A ready-to-import instance with the right Minecraft and loader version. Download the
          zip, or copy the URL and paste it into "Add Instance" &rarr; "Import". Both contain
          {{ pack.isPublic ? 'the public' : 'your personal' }} link{{ pack.isPublic ? '' : ', so do not share them' }}.
        </p>
        <div class="d-flex flex-wrap ga-2 mb-6">
          <v-btn
            variant="tonal"
            density="comfortable"
            prepend-icon="mdi-download"
            text="Download instance (.zip)"
            :loading="downloading"
            @click="downloadInstance"
          />
          <v-btn
            variant="tonal"
            density="comfortable"
            prepend-icon="mdi-clipboard-text-multiple-outline"
            text="Copy import URL"
            @click="copyInstanceUrl"
          />
        </div>

        <h4 class="mb-1">
          Manual client setup
        </h4>
        <p class="text-medium-emphasis mb-3">
          Paste this command into the instance's custom commands as a pre-launch command.
          It keeps the instance in sync with this pack.
        </p>
        <div class="d-flex flex-column ga-3 align-start">
          <v-btn
            variant="tonal"
            density="comfortable"
            prepend-icon="mdi-console"
            text="Copy setup command"
            @click="copySetupCommand"
          />
          <a
            href="https://packwiz.infra.link/tutorials/installing/packwiz-installer/"
            target="_blank"
            rel="noopener noreferrer"
          >
            packwiz-installer guide
          </a>
        </div>
      </v-card-text>

      <v-card-actions class="justify-end">
        <v-btn
          variant="text"
          text="Close"
          @click="model = false"
        />
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>
