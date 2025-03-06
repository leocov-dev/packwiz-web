<script setup lang="ts">
import {Pack, PackPermission} from "@/interfaces/pack.ts";
import PackActions from "@/components/PackActions.vue";

const {pack} = defineProps<{ pack: Pack }>()

const permissionMap = {
  [PackPermission.STATIC]: "-",
  [PackPermission.VIEW]: "View",
  [PackPermission.EDIT]: "Edit",
}
</script>

<template>
  <v-card min-height="120">
    <v-card-title class="d-flex justify-center">
      <h4 class="me-auto">
        {{ pack.title }}
      </h4>
      <PackStatus :status="pack.isArchived ? 'archived' : pack.status"/>

      <PackStatus
        v-if="pack.isPublic"
        class="ms-2"
        status="public"
      />

      <PackStatus
        v-if="pack.dataMissing"
        class="ms-2"
        status="warning"
      />

    </v-card-title>
    <v-card-text>
      {{ pack.description }}
    </v-card-text>
    <v-card-actions class="ms-2 me-2 d-flex justify-end">
      <v-btn
        v-if="pack.permission >= PackPermission.VIEW"
        class="me-auto"
        :text="permissionMap[pack.permission]"
        :to="`/packs/${pack.slug}`"
        variant="tonal"
        density="comfortable"
      />

      <PackActions :pack="pack"/>
    </v-card-actions>
  </v-card>
</template>
