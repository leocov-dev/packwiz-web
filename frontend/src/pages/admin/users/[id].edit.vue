<route lang="yaml">
meta:
  layout: app
</route>

<script setup lang="ts">
import {useRoute} from "vue-router";

const route = useRoute<'/admin/users/[id]'>()
</script>

<template>
  {{ route.params.id }}

  <UnderConstruction />
</template>
