<script setup lang="ts">
import {type Pack, PackStatus} from "@/interfaces/pack.ts";
import PackLinksDialog from "@/components/pack/PackLinksDialog.vue";
import {publicPackPath} from "@/services/packs.service.ts";

const {pack} = defineProps<{ pack: Pack }>()

const showDialog = ref(false)

const actionsDisabled = computed(() => {
  return pack.isArchived || pack.status === PackStatus.DRAFT
})

const title = "Links and setup"
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
          density="comfortable"
          variant="text"
          icon="mdi-link-variant"
          :aria-label="title"
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
      :text="title"
      location="bottom"
    >
      <template #activator="{ props }">
        <v-btn
          v-bind="props"
          density="comfortable"
          variant="text"
          icon="mdi-link-variant"
          :aria-label="title"
          @click="showDialog = true"
        />
      </template>
    </v-tooltip>

    <v-tooltip
      v-if="pack.isPublic"
      text="Open public page"
      location="bottom"
    >
      <template #activator="{ props }">
        <v-btn
          v-bind="props"
          density="comfortable"
          variant="text"
          icon="mdi-open-in-new"
          aria-label="Open public page"
          :href="publicPackPath(pack.slug)"
          target="_blank"
          rel="noopener"
        />
      </template>
    </v-tooltip>

    <PackLinksDialog
      v-model="showDialog"
      :pack="pack"
    />
  </div>
</template>
