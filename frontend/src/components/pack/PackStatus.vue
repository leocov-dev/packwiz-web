<script setup lang="ts">

const { status, class: classDef } = defineProps<{
  status: "draft" | "published" | "public" | "archived" | "warning",
  class?: string
}>()

const statusMap = {
  published: {
    icon: 'mdi-earth',
    tooltip: 'This pack is published',
    color: 'primary',
  },
  draft: {
    icon: 'mdi-file-edit',
    tooltip: 'This pack is a draft',
    color: 'secondary',
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
  },
  warning: {
    icon: 'mdi-alert',
    tooltip: 'This pack has invalid file data',
    color: 'warning',
  }
}

</script>

<template>
  <v-tooltip :text="statusMap[status].tooltip">
    <template #activator="{ props }">
      <v-icon
        v-bind="props"
        :icon="statusMap[status].icon"
        :color="statusMap[status].color"
        :class="classDef"
      />
    </template>
  </v-tooltip>
</template>

