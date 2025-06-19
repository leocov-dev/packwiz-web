<script setup lang="ts">

import type {Mod} from "@/interfaces/pack.ts";

const {packId, mod} = defineProps<{ packId: number, mod: Mod }>()


const modTypeIconMap: {[key: string]: string} = {
  "mods": "mdi-shield-sword-outline",
  "resourcepacks": "mdi-package-variant-closed",
  "shaderpacks": "mdi-crystal-ball",
  "plugins": "mdi-power-socket-us",
}


const openLink = () => {
  // window.open(mod.sourceLink, '_blank')
}

</script>

<template>
  <v-card class="ma-1 ps-5 pe-5 pt-3 pb-3 elevation-4">
    <div class="d-flex align-center">
      <v-icon
        v-tooltip="mod.type"
        class="me-2"
        :icon="modTypeIconMap[mod.type] || 'mdi-puzzle-outline'"
      />

      <div>
        {{ mod.name }}
      </div>
      <!--      <v-btn-->
      <!--        v-if="mod.sourceLink"-->
      <!--        class="ms-2"-->
      <!--        link-->
      <!--        density="comfortable"-->
      <!--        color="default"-->
      <!--        variant="plain"-->
      <!--        icon="mdi-open-in-new"-->
      <!--        @click="openLink"-->
      <!--      />-->

      <v-spacer />

      <div
        class="ms-4 me-8 text-subtitle-2 text-disabled text-truncate"
      >
        {{ mod.fileName }}
      </div>

      <div class="d-flex justify-end">
        <v-icon
          v-if="mod.side === 'client'"
          v-tooltip="'client'"
          class="me-2"
          icon="mdi-account-outline"
        />
        <v-icon
          v-if="mod.side === 'server'"
          v-tooltip="'server'"
          class="me-2"
          icon="mdi-server-outline"
        />
        <v-icon
          v-if="mod.side === 'both'"
          v-tooltip="'server+client'"
          class="me-2"
          icon="mdi-circle-double"
        />
        <v-icon
          v-tooltip="mod.pinned ? 'pinned' : 'unpinned'"
          class="me-2"
          :icon="mod.pinned ? 'mdi-pin' : 'mdi-pin-off-outline'"
        />

        <v-btn
          density="comfortable"
          color="warning"
          variant="outlined"
          text="Edit"
          :to="`${packId}/mod/${mod.id}`"
        />
      </div>
    </div>
  </v-card>
</template>
