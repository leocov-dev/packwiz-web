<script setup lang="ts">
import {PackResponse} from "@/interfaces/pack.ts";
import PackActions from "@/components/pack/PackActions.vue";
import {usePackPermissions} from "@/composables/usePackPermissions.ts";
import {Perm} from "@/lib/permissions.ts";
import {packStatusKey} from "@/lib/pack-status.ts";
import {toTitleCase} from "@/services/utils.ts";

const loaderColors: Record<string, string> = {
  fabric: 'brown',
  forge: 'blue-grey',
  neoforge: 'orange',
  quilt: 'purple',
  liteloader: 'teal',
}

const {pack} = defineProps<{ pack: PackResponse }>()

const {can} = usePackPermissions(() => pack)

const loaderColor = computed(() => loaderColors[pack.loader?.toLowerCase()] ?? 'grey')

const isDraft = computed(() => pack.status === 'draft' && !pack.isArchived)

const viewLabel = computed(() => can(Perm.PackInfoEdit) ? "Edit" : "View")
</script>

<template>
  <v-card
    min-height="160"
    max-height="160"
    class="d-flex flex-column"
    :class="{ 'pack-card--draft': isDraft }"
  >
    <v-card-title class="d-flex justify-center">
      <h4 class="me-auto">
        {{ pack.name }}
      </h4>
      <PackStatus :status="packStatusKey(pack)" />
    </v-card-title>

    <v-card-text class="multiline-truncate flex-grow-1">
      {{ pack.description }}
    </v-card-text>

    <v-card-actions class="ms-2 me-2 d-flex justify-end">
      <v-btn
        v-if="can(Perm.PackView)"
        class="me-3"
        :text="viewLabel"
        :to="`/packs/${pack.id}`"
        variant="tonal"
        density="comfortable"
      />

      <v-chip
        v-if="pack.loader"
        :color="loaderColor"
        variant="flat"
        size="small"
        label
      >
        {{ toTitleCase(pack.loader) }}
      </v-chip>

      <v-chip
        v-if="pack.mcVersion"
        color="green-darken-2"
        variant="flat"
        size="small"
        label
        class="ms-2"
      >
        {{ pack.mcVersion }}
      </v-chip>

      <v-spacer />

      <PackActions :pack="pack" />
    </v-card-actions>
  </v-card>
</template>

<style scoped>
/* drafts: slightly greyer, less saturated surface */
.pack-card--draft {
  background-color: color-mix(in srgb, rgb(var(--v-theme-surface)) 85%, grey);
}

/* No Vuetify utility clamps to N lines (see STYLE_GUIDE.md) */
.multiline-truncate {
  display: -webkit-box;
  line-clamp: 2;
  -webkit-line-clamp: 2; /* Number of lines to show */
  -webkit-box-orient: vertical;
  overflow: hidden;
  white-space: pre-wrap; /* This preserves new lines */
}
</style>
