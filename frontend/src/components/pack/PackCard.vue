<script setup lang="ts">
import {PackResponse} from "@/interfaces/pack.ts";
import PackActions from "@/components/pack/PackActions.vue";
import {usePackPermissions} from "@/composables/usePackPermissions.ts";
import {Perm} from "@/lib/permissions.ts";

const {pack} = defineProps<{ pack: PackResponse }>()

const {can} = usePackPermissions(() => pack)

const viewLabel = computed(() => can(Perm.PackInfoEdit) ? "Edit" : "View")
</script>

<template>
  <v-card
    min-height="160"
    max-height="160"
    class="d-flex flex-column"
  >
    <v-card-title class="d-flex justify-center">
      <h4 class="me-auto">
        {{ pack.name }}
      </h4>
      <PackStatus :status="pack.isArchived ? 'archived' : pack.status" />

      <PackStatus
        v-if="pack.isPublic"
        class="ms-2"
        status="public"
      />
    </v-card-title>

    <v-card-text class="multiline-truncate flex-grow-1">
      {{ pack.description }}
    </v-card-text>

    <v-card-actions class="ms-2 me-2 d-flex justify-end">
      <v-btn
        v-if="can(Perm.PackView)"
        class="me-auto"
        :text="viewLabel"
        :to="`/packs/${pack.id}`"
        variant="tonal"
        density="comfortable"
      />

      <PackActions :pack="pack" />
    </v-card-actions>
  </v-card>
</template>

<style scoped>
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
