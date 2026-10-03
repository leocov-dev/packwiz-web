<script setup lang="ts">
import type {Pack} from "@/interfaces/pack.ts";
import {
  clientSetupCommandToClipboard,
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
          Client setup (MultiMC / Prism)
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
