<template>
  <v-footer
    elevation="5"
    height="40"
  >
    <div
      class="text-caption text-disabled"
      style="position: absolute; right: 16px;"
    />
  </v-footer>
</template>

<script setup lang="ts">
</script>

<style scoped lang="sass">
  .social-link :deep(.v-icon)
    color: rgba(var(--v-theme-on-background), var(--v-disabled-opacity))
    text-decoration: none
    transition: .2s ease-in-out

    &:hover
      color: rgba(25, 118, 210, 1)
</style>
