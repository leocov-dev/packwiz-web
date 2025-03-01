<script setup lang="ts">

const { status } = defineProps<{ status: "draft" | "published" | "public" | "archived" }>()

const statusMap = {
  published: {
    icon: 'mdi-earth',
    tooltip: 'This pack is published',
    color: 'primary',
  },
  draft: {
    icon: 'mdi-file-edit',
    tooltip: 'This pack is a draft',
    color: 'warning',
  },
  archived: {
    icon: 'mdi-archive-outline',
    tooltip: 'This pack is archived',
    color: 'disabled',
  },
  public: {
    icon: 'mdi-account-group',
    tooltip: 'This pack does not require authentication',
    color: 'red',
  }
}

</script>

<template>
  <v-tooltip :text="statusMap[status].tooltip">
    <template v-slot:activator="{ props }">
      <v-icon
        v-bind="props"
        :icon="statusMap[status].icon"
        :color="statusMap[status].color"
      ></v-icon>
    </template>
  </v-tooltip>
</template>

