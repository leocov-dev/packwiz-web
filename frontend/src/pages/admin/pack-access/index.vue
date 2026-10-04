<route lang="yaml">
meta:
  layout: app
</route>

<script setup lang="ts">
import SystemPackAccess from "@/components/access/SystemPackAccess.vue";

const tab = ref<'syncs' | 'downloads'>('syncs')
</script>

<template>
  <div class="ma-6">
    <v-tabs
      v-model="tab"
      class="mb-4"
      color="primary"
    >
      <v-tab
        value="syncs"
        prepend-icon="mdi-sync"
        text="Pack syncs"
      />
      <v-tab
        value="downloads"
        prepend-icon="mdi-download"
        text="Instance downloads"
      />
    </v-tabs>

    <v-window v-model="tab">
      <v-window-item value="syncs">
        <SystemPackAccess />
      </v-window-item>
      <v-window-item value="downloads">
        <SystemPackAccess
          source="instance-downloads"
          title="Instance Downloads"
        />
      </v-window-item>
    </v-window>
  </div>
</template>
