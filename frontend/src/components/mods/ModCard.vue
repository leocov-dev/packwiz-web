<script setup lang="ts">

import type {ModData} from "@/interfaces/pack.ts";

const {mod} = defineProps<{ mod: ModData }>()

const openLink = () => {
  window.open(mod.sourceLink, '_blank')
}

</script>

<template>
  <v-card class="ma-1 ps-5 pe-5 pt-3 pb-3 elevation-4">
    <div class="d-flex align-center">
      <div>
        {{ mod.displayName }}
      </div>
      <v-btn
        v-if="mod.sourceLink"
        class="ms-2"
        link
        density="comfortable"
        color="default"
        variant="plain"
        icon="mdi-open-in-new"
        @click="openLink"
      />

      <v-spacer />

      <div
        class="ms-4 me-8 text-subtitle-2 text-disabled text-truncate"
      >
        {{ mod.filename }}
      </div>

      <div class="d-flex justify-end">
        <v-icon
          v-if="mod.side === 'client' || mod.side === 'both'"
          class="me-2"
          icon="mdi-account-outline"
        />
        <v-icon
          v-if="mod.side === 'server' || mod.side === 'both'"
          class="me-2"
          icon="mdi-server-outline"
        />
        <v-icon
          class="me-2"
          :icon="mod.pinned ? 'mdi-pin' : 'mdi-pin-off-outline'"
        />

        <v-btn
          density="comfortable"
          color="warning"
          variant="outlined"
          text="Edit"
        />
      </div>
    </div>
  </v-card>
</template>
